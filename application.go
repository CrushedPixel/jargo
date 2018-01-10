package jargo

import (
	"errors"
	"fmt"
	"github.com/crushedpixel/jargo/internal"
	"github.com/crushedpixel/margo"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg"
	"gopkg.in/go-playground/validator.v9"
	"log"
	"net/url"
	"reflect"
)

// DefaultMaxPageSize is the default maximum number
// of allowed entries per page.
const DefaultMaxPageSize = 25

// DefaultErrorHandler is the default error handler.
// It sends errors to the client using NewErrorResponse.
// If the recovered value is not an error, it sends
// ErrInternalServerError.
func DefaultErrorHandler(c *gin.Context, r interface{}) {
	var err error
	var ok bool

	err, ok = r.(error)
	if !ok {
		log.Println(fmt.Sprintf("%v", r))
		err = ErrInternalServerError
	}

	res := NewErrorResponse(err)
	err = res.Send(c)
	if err != nil {
		panic(err)
	}
}

// Application is the central component of jargo.
type Application struct {
	*margo.Application

	db *pg.DB

	registry  internal.SchemaRegistry
	resources map[*internal.Schema]*Resource

	controllers []*Controller
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
	server := margo.NewApplication()
	server.ErrorHandler = DefaultErrorHandler
	return &Application{
		Application: server,
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

// AddController registers a Controller's Actions
// with the Application.
func (app *Application) AddController(c *Controller) {
	for route, action := range c.actions {
		if action == nil {
			continue
		}

		// generate endpoint path
		path := c.Namespace()
		// append forward slash to path
		// if namespace doesn't end with one
		if len(path) < 1 || path[len(path)-1] != '/' {
			path += "/"
		}
		// append Resource member name
		path += c.Resource().JSONAPIName()
		// append forward slash to path
		// if route path doesn't start with one
		if len(route.path) < 1 || path[0] != '/' {
			path += "/"
		}
		// append route path to path
		path += route.path

		// register an endpoint for each action
		app.Endpoint(margo.NewEndpoint(
			route.method,
			path,
			// first, call controller-level middleware
			HandlerChain(c.handlers(app)).ToMargoHandler(),
			// call action handlers
			action.Handlers().ToMargoHandler()))
	}
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

// ParsePagination parses pagination parameters
// from an URL query according to the JSON API spec.
// See http://jsonapi.org/format/#fetching-pagination
func (app *Application) ParsePagination(query url.Values) (*Pagination, error) {
	return parsePagination(query, app.maxPageSize)
}

// Run starts the Application,
// serving HTTP requests on the specified address.
func (app *Application) Run(addr ...string) error {
	return app.Application.Run(addr...)
}
