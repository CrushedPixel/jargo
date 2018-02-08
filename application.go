package jargo

import (
	"errors"
	"github.com/crushedpixel/jargo/internal"
	"github.com/go-pg/pg"
	"gopkg.in/go-playground/validator.v9"
	"reflect"
)

// DefaultMaxPageSize is the default maximum number
// of allowed entries per page.
const DefaultMaxPageSize = 25

var errNoResponse = errors.New("the last HandlerFunc returned a nil Response")

// Application is the central component of jargo.
type Application struct {
	db *pg.DB

	controllers map[*Resource]*Controller

	registry  internal.SchemaRegistry
	resources map[*internal.Schema]*Resource

	maxPageSize int
	validate    *validator.Validate
}

// NewApplication returns a new Application
// using the given pg handle and a default
// Validate instance.
func NewApplication(db *pg.DB) *Application {
	return NewApplicationWithValidate(db, validator.New())
}

// NewApplicationWithErrorHandler returns a new Application
// using the given pg handle and Validate instance.
func NewApplicationWithValidate(db *pg.DB, validate *validator.Validate) *Application {
	return &Application{
		registry:    make(internal.SchemaRegistry),
		resources:   make(map[*internal.Schema]*Resource),
		db:          db,
		maxPageSize: DefaultMaxPageSize,
		validate:    validate,
	}
}

// DB returns the pg database handle used
// by the Application.
func (app *Application) DB() *pg.DB {
	return app.db
}

// Validate returns the Validate instance
// used to validate create and update requests.
func (app *Application) Validate() *validator.Validate {
	return app.validate
}

// RegisterResource registers and initializes a Resource
// and all related Resources.
// If the Resource has already been registered,
// its cached value is returned.
//
// Panics if model is not an instance of a properly annotated
// Resource Model.
func (app *Application) RegisterResource(model interface{}) (*Resource, error) {
	s, err := app.registry.RegisterSchema(reflect.TypeOf(model))
	if err != nil {
		return nil, err
	}

	if resource, ok := app.resources[s]; ok {
		return resource, nil
	}

	for _, schema := range app.registry {
		if _, ok := app.resources[schema]; !ok {
			resource := &Resource{schema: schema}
			err := resource.Initialize(app.DB())
			if err != nil {
				return nil, err
			}

			app.resources[schema] = resource
		}
	}

	return app.resources[s], nil
}

// MustRegisterResource calls RegisterResource
// and panics if it encounters an error.
func (app *Application) MustRegisterResource(model interface{}) *Resource {
	r, err := app.RegisterResource(model)
	if err != nil {
		panic(err)
	}
	return r
}

func (app *Application) AddController(c *Controller) {
	app.controllers[c.resource] = c
}

func (app *Application) Controllers() map[*Resource]*Controller {
	return app.controllers
}

// MaxPageSize returns the maximum number
// of allowed entries per page for paginated results.
func (app *Application) MaxPageSize() int {
	return app.maxPageSize
}

// SetMaxPageSize sets the maximum number
// of allowed entries per page.
// Panics if value is not positive.
func (app *Application) SetMaxPageSize(maxPageSize int) {
	if maxPageSize < 1 {
		panic(errors.New("maximum page size has to be positive"))
	}
	app.maxPageSize = maxPageSize
}

// Handle handles a request and returns response http status and payload.
func (app *Application) Handle(request *Request) (int, string) {
	res := app.handle(request)
	payload, err := res.Payload()
	if err != nil {
		// handle error on payload creation
		println("jargo: Error resolving payload", err.Error()) // TODO: proper logging
		if p, ok := err.(Response); ok {
			res = p
		} else {
			res = ErrInternalServerError
		}
		payload, _ = res.Payload()
	}

	return res.Status(), payload
}

func (app *Application) handle(request *Request) Response {
	// get Resource for JSON API resource name
	var resource *Resource
	for _, r := range app.resources {
		if r.JSONAPIName() == request.ResourceName {
			resource = r
			break
		}
	}
	if resource == nil {
		return ErrNotFound
	}

	// get controller for request resource
	controller, ok := app.controllers[resource]
	if !ok {
		return ErrNotFound
	}
	// get controller action handlers for requested action
	handlers, ok := controller.Actions[request.ActionType]
	if !ok {
		return ErrNotFound
	}

	// create context object
	context, err := NewContext(app, resource, request)
	if err != nil {
		return NewErrorResponse(err)
	}

	// prepend controller middleware to action handlers
	allHandlers := append(controller.middleware, handlers...)
	// execute action handlers
	var response Response
	for _, handler := range allHandlers {
		response = handler(context)
		if response != nil {
			return response
		}
	}

	return NewErrorResponse(errNoResponse)
}
