package jargo

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/crushedpixel/cement"
	"github.com/crushedpixel/jargo/internal"
	"github.com/desertbit/glue"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/google/jsonapi"
	"github.com/json-iterator/go"
	"net/http"
	"sync"
	"time"
)

const (
	triggerFunctionName     = "jargo_realtime_notify"
	notificationChannelName = "jargo_realtime"
)

var (
	errAlreadyRunning = errors.New("realtime instance is already running")
	errNotRunning     = errors.New("realtime instance must be started to be able to handle http requests")
)

// triggerFunctionQuery is a query creating a trigger function
// that notifies the notification channel whenever a row is inserted, updated or deleted.
var triggerFunctionQuery = fmt.Sprintf(`
CREATE OR REPLACE FUNCTION %s() RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    PERFORM pg_notify('%s', json_build_object('table', TG_TABLE_NAME,
      'id', NEW.id::text, 'type', TG_OP,
      'new', row_to_json(NEW)::text
    )::text);
  ELSIF TG_OP = 'DELETE' THEN
    PERFORM pg_notify('%s', json_build_object('table', TG_TABLE_NAME,
      'id', OLD.id::text, 'type', TG_OP,
      'old', row_to_json(OLD)::text
    )::text);
  ELSIF TG_OP = 'UPDATE' THEN
    PERFORM pg_notify('%s', json_build_object('table', TG_TABLE_NAME,
      'id', OLD.id::text, 'type', TG_OP,
      'old', row_to_json(OLD)::text,
      'new', row_to_json(NEW)::text
    )::text);
  ELSE
    RAISE EXCEPTION 'Invalid Trigger Operation: %%', TG_OP;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
`,
	triggerFunctionName,
	notificationChannelName,
	notificationChannelName,
	notificationChannelName)

// notificationPayload is a struct representation
// of the json payload sent to jargo_realtime listeners
type notificationPayload struct {
	// The table name of the modified record
	Table string `json:"table"`
	// The id of the modified record
	Id string `json:"id"`
	// Type of the action made to the row
	Type string `json:"type"`
	// The original record, json-encoded
	OldRecord string `json:"old"`
	// The modified record, json-encoded
	NewRecord string `json:"new"`
}

const (
	// msgConnectionTimeout is sent to the client
	// if they don't send a connection message
	// before Realtime.ConnectionMessageTimeout is exceeded.
	msgConnectionTimeout = "CONNECTION_TIMEOUT"

	// msgConnectionDisallowed is sent to the client
	// if Realtime.HandleConnection returns false.
	msgConnectionDisallowed = "CONNECTION_DISALLOWED"

	subscribeChannelName = "subscribe"
	deletedChannelName   = "deleted"
	updatedChannelName   = "updated"

	msgOk              = `{"status":"ok"}`
	msgInvalidResource = `{"error":"INVALID_RESOURCE"}`
	msgAccessDenied    = `{"error":"ACCESS_DENIED"}`
)

// subscribePayload is a struct representing the JSON payload
// sent by Realtime clients when subscribing to a resource.
type subscribePayload struct {
	// Model is the JSON API member name of the resource type to subscribe to
	Model string `json:"model"`
	// Id is the id of the resource to subscribe to
	Id int64 `json:"id,string"`
}

// resourceDeletedPayload is a struct representing the JSON payload
// to send to Realtime clients when a resource was deleted.
type resourceDeletedPayload struct {
	Model string `json:"model"`
	Id    string `json:"id"`
}

// resourceUpdatedPayload is a struct representing the JSON payload
// to send to Realtime clients when a resource was inserted or updated.
type resourceUpdatedPayload struct {
	Model   string `json:"model"`
	Id      string `json:"id"`
	Payload string `json:"payload"`
}

// Realtime allows clients to subscribe to
// resource instances via websocket.
type Realtime struct {
	*glue.Server

	app *Application

	// namespace is the namespace on which to listen for requests.
	namespace string

	// ConnectionMessageTimeout is the time
	// to wait for the connection message.
	// Defaults to 10s.
	ConnectionMessageTimeout time.Duration

	// HandleConnection is is the HandleConnectionFunc
	// to be invoked when a new socket connection
	// has sent their connection message.
	// If it returns true, the connection is allowed, otherwise
	// it is immediately closed.
	// Defaults to a function always returning true.
	HandleConnection HandleConnectionFunc

	MaySubscribe MaySubscribeFunc

	// subscriptions is a map containing all subscriptions
	// for a socket.
	subscriptions map[*glue.Socket]map[*Resource][]int64
	// subscriptionsMutex is the mutex protecting subscriptions
	subscriptionsMutex *sync.Mutex

	// connectingSockets is the channel to which
	// all sockets that have just connected are written.
	connectingSockets chan *glue.Socket

	// release is the channel that signals internal goroutines
	// to finish execution when closed
	release chan *struct{}

	// running indicates whether the Realtime instance
	// is currently able to handle requests.
	running bool
}

type HandleConnectionFunc func(socket *glue.Socket, message string) bool
type MaySubscribeFunc func(socket *glue.Socket, resource *Resource, id int64) bool

func defaultHandleConnectionFunc(*glue.Socket, string) bool {
	return true
}

func defaultMaySubscribeFunc(*glue.Socket, *Resource, int64) bool {
	return true
}

// NewRealtime returns a new Realtime instance for an Application and namespace
// using the default HandleConnection and MaySubscribe handlers,
// which allow all connections and subscriptions.
func NewRealtime(app *Application, namespace string) *Realtime {
	r := &Realtime{
		app: app,

		ConnectionMessageTimeout: 10 * time.Second,
		HandleConnection:         defaultHandleConnectionFunc,

		MaySubscribe: defaultMaySubscribeFunc,

		connectingSockets: make(chan *glue.Socket, 0),

		subscriptions:      make(map[*glue.Socket]map[*Resource][]int64),
		subscriptionsMutex: &sync.Mutex{},
	}
	r.SetNamespace(namespace)
	return r
}

// SetNamespace sets the namespace the realtime instance listens on.
// Panics if the realtime instance is running.
func (r *Realtime) SetNamespace(namespace string) {
	if r.running {
		panic(errAlreadyRunning)
	}
	r.namespace = NormalizeNamespace(namespace)
}

// Namespace returns the namespace the realtime instance listens on.
func (r *Realtime) Namespace() string {
	return r.namespace
}

func (r *Realtime) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !r.running {
		panic(errNotRunning)
	}
	r.Server.ServeHTTP(w, req)
}

// Bridge registers the Realtime instance with a ServeMux.
// It prepares the Realtime instance for request handling by invoking Start.
func (r *Realtime) Bridge(mux *http.ServeMux) error {
	err := r.Start()
	if err != nil {
		return err
	}
	mux.Handle(r.namespace, r)
	return nil
}

// Start prepares the Realtime instance to handle incoming requests.
// This must be called before registering the Realtime instance
// as an http handler.
//
// After handling is done, Release should be called to stop all internal
// goroutines.
func (r *Realtime) Start() error {
	if r.running {
		return errAlreadyRunning
	}

	r.running = true
	r.release = make(chan *struct{}, 0)

	// initialize glue server
	s := glue.NewServer(glue.Options{
		HTTPHandleURL: r.namespace,
		CheckOrigin: func(r *http.Request) bool {
			return true // TODO add a way to specify a CheckOrigin function?
		},
	})

	s.OnNewSocket(r.onNewSocket)
	r.Server = s

	// create trigger function calling notify
	_, err := r.app.DB().Exec(triggerFunctionQuery)
	if err != nil {
		return err
	}

	// create triggers on database tables
	for _, resource := range r.app.resources {
		err := resource.schema.CreateRealtimeTriggers(r.app.DB(), triggerFunctionName)
		if err != nil {
			return err
		}
	}

	// create notification channel for database trigger
	notificationChannel := r.app.DB().Listen(notificationChannelName).Channel()

	go r.handleConnectingSockets()
	go r.handleRowUpdates(notificationChannel)
	return nil
}

// Release stops all internal goroutines.
// Should be called after serving is done.
func (r *Realtime) Release() {
	r.Server.Release()
	close(r.release)
	r.running = false
}

func (r *Realtime) onNewSocket(s *glue.Socket) {
	r.connectingSockets <- s
}

func (r *Realtime) handleConnectingSockets() {
	for {
		select {
		case socket := <-r.connectingSockets:
			message, err := socket.Read(r.ConnectionMessageTimeout)
			if err != nil {
				if err != glue.ErrSocketClosed {
					// no connection message received
					socket.Write(msgConnectionTimeout)
					socket.Close()
				}
				break
			}
			if !r.HandleConnection(socket, message) {
				// connection disallowed
				socket.Write(msgConnectionDisallowed)
				socket.Close()
				break
			}

			r.initSocketConnection(socket)
		case <-r.release:
			// closing the release channel escapes the for loop
			return
		}
	}
}

func (r *Realtime) handleRowUpdates(channel <-chan *pg.Notification) {
	for {
		select {
		case notification := <-channel:
			payload := &notificationPayload{}
			err := json.Unmarshal([]byte(notification.Payload), payload)
			if err != nil {
				panic(err)
			}

			var resource *Resource
			for _, res := range r.app.resources {
				if res.schema.Table() == payload.Table {
					resource = res
					break
				}
			}
			if resource == nil {
				panic(errors.New("resource for table name not found"))
			}

			// map of all resources affected by the change
			updated := make(map[*Resource][]interface{})

			// add all resources that were updated by the resources'
			// relationships being modified
			var modifiedInstance *internal.SchemaInstance
			if payload.Type == "INSERT" || payload.Type == "DELETE" {
				var recordPayload string
				if payload.Type == "INSERT" {
					recordPayload = payload.NewRecord
				} else {
					recordPayload = payload.OldRecord
				}

				instance, err := decodeJsonRecord(resource, recordPayload)
				if err != nil {
					panic(err)
				}

				// get all resources this resource was/is now
				// related to and add them to the updated map
				modifiedInstance = resource.schema.ParsePGModel(instance)
				for schema, ids := range modifiedInstance.GetRelationIds() {
					// get resource for schema
					for _, res := range r.app.resources {
						if res.schema == schema {
							updated[res] = ids
						}
					}
				}
			} else if payload.Type == "UPDATE" {
				oldInstance, err := decodeJsonRecord(resource, payload.OldRecord)
				if err != nil {
					panic(err)
				}
				oldRelations := resource.schema.ParsePGModel(oldInstance).GetRelationIds()

				newInstance, err := decodeJsonRecord(resource, payload.NewRecord)
				if err != nil {
					panic(err)
				}
				modifiedInstance = resource.schema.ParsePGModel(newInstance)
				newRelations := modifiedInstance.GetRelationIds()

				// create a changeset of all of the resources' relations
				changeset := make(map[*internal.Schema][]interface{})

				// get all relations that were removed
				for schema, oldIds := range oldRelations {
					newIds := newRelations[schema]
					removed := difference(oldIds, newIds)
					changeset[schema] = append(changeset[schema], removed...)
				}

				// get all relations that were added
				for schema, newIds := range newRelations {
					oldIds := oldRelations[schema]
					added := difference(newIds, oldIds)
					changeset[schema] = append(changeset[schema], added...)
				}

				// apply changeset to updated map
				for schema, ids := range changeset {
					// get resource for schema
					for _, res := range r.app.resources {
						if res.schema == schema {
							updated[res] = ids
						}
					}
				}
			} else {
				panic(errors.New("unknown trigger event type"))
			}

			if payload.Type == "DELETE" {
				sockets := r.subscribers(resource, payload.Id)
				if len(sockets) > 0 {
					sendResourceDeleted(sockets, resource, payload.Id)
				}
			} else {
				// add resource to map of resources that were updated
				updated[resource] = append(updated[resource], payload.Id)
			}

			// send updates for all updated resources to subscribed clients
			for resource, ids := range updated {
				for _, id := range ids {
					sockets := r.subscribers(resource, id)
					if len(sockets) > 0 {
						// fetch updated resource instance from database
						m, err := resource.SelectById(r.app.DB(), id).Result()
						if err != nil {
							panic(err)
						}
						instance := resource.schema.ParseResourceModel(m)
						err = sendResourceUpdated(sockets, resource, payload.Id, instance)
						if err != nil {
							panic(err)
						}
					}
				}
			}

			break
		case <-r.release:
			// closing the release channel escapes the for loop
			return
		}
	}
}

// decodeJsonRecord parses a json-encoded record into a pg model instance
func decodeJsonRecord(resource *Resource, payload string) (interface{}, error) {
	instance := resource.schema.NewPGModelInstance()
	model, err := orm.NewModel(instance)
	if err != nil {
		return nil, err
	}

	it := jsoniter.ParseString(jsoniter.ConfigDefault, payload)
	record := it.ReadAny()
	for i, key := range record.Keys() {
		model.ScanColumn(i, key, []byte(record.Get(key).ToString()))
	}

	return instance, nil
}

func (r *Realtime) initSocketConnection(socket *glue.Socket) {
	subscribeChannel := socket.Channel(subscribeChannelName)
	subscribeChannel.OnRead(cement.Glue(subscribeChannel, r.onSubscribeRead))
}

func (r *Realtime) onSubscribeRead(socket *glue.Socket, messageId string, data string) (int, string) {
	payload := &subscribePayload{}
	err := json.Unmarshal([]byte(data), payload)
	if err != nil {
		return cement.CodeError, cement.MsgInvalidPayload
	}

	// get resource with from application's registry
	var resource *Resource
	for schema, r := range r.app.resources {
		if payload.Model == schema.JSONAPIName() {
			resource = r
			break
		}
	}
	if resource == nil {
		return cement.CodeError, msgInvalidResource
	}

	// call MaySubscribe hook
	if !r.MaySubscribe(socket, resource, payload.Id) {
		return cement.CodeError, msgAccessDenied
	}

	// subscribe client to resource
	r.subscriptionsMutex.Lock()
	s, ok := r.subscriptions[socket]
	if !ok {
		s = make(map[*Resource][]int64)
		r.subscriptions[socket] = s
	}

	s[resource] = append(s[resource], payload.Id)
	r.subscriptionsMutex.Unlock()

	return cement.CodeOk, msgOk
}

// subscribers returns all sockets that are subscribed to a resource instance.
func (r *Realtime) subscribers(resource *Resource, id interface{}) []*glue.Socket {
	var sockets []*glue.Socket
	r.subscriptionsMutex.Lock()
	for socket, subscriptions := range r.subscriptions {
		for _, i := range subscriptions[resource] {
			if id == i {
				sockets = append(sockets, socket)
			}
		}
	}
	r.subscriptionsMutex.Unlock()
	return sockets
}

func sendResourceDeleted(sockets []*glue.Socket, resource *Resource, id string) error {
	b, err := jsoniter.ConfigDefault.Marshal(&resourceDeletedPayload{
		Model: resource.JSONAPIName(),
		Id:    id,
	})
	if err != nil {
		return err
	}
	str := string(b)
	for _, socket := range sockets {
		channel := socket.Channel(deletedChannelName)
		channel.DiscardRead()
		channel.Write(str)
	}
	return nil
}

func sendResourceUpdated(sockets []*glue.Socket, resource *Resource, id string, instance *internal.SchemaInstance) error {
	p, err := jsonapi.Marshal(instance.ToJsonapiModel())
	if err != nil {
		return err
	}
	payload := p.(*jsonapi.OnePayload)
	payload.Included = nil

	resourceBytes, err := jsoniter.ConfigDefault.Marshal(p)
	if err != nil {
		return err
	}

	b, err := jsoniter.ConfigDefault.Marshal(&resourceUpdatedPayload{
		Model:   resource.JSONAPIName(),
		Id:      id,
		Payload: string(resourceBytes),
	})
	if err != nil {
		return err
	}
	str := string(b)
	for _, socket := range sockets {
		channel := socket.Channel(updatedChannelName)
		channel.DiscardRead()
		channel.Write(str)
	}
	return nil
}
