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

// Application is the central component of jargo.
type Application struct {
	db *pg.DB

	registry  internal.SchemaRegistry
	resources map[*internal.Schema]*Resource

	controllers map[*Resource]*Controller
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

// AddController registers a Controller's Actions
// with the Application.
func (app *Application) AddController(c *Controller) {

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

// Handle handles a request and returns a response.
func (app *Application) Handle(request Request) Response {
	// get Resource for JSON API resource name
	var resource *Resource
	for _, r := range app.resources {
		if r.JSONAPIName() == request.Resource() {
			resource = r
			break
		}
	}

	if resource == nil {
		return ErrNotFound
	}

	// based on request's action type,
	// forward request to action handler
	controller, ok := app.controllers[resource]
	if !ok {
		return ErrNotFound
	}

	context := &Context{
		application: app,
		resource:    resource,
		data:        make(map[string]interface{}),
	}
	return controller.Handle(context, request)
}
