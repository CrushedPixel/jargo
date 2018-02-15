package jargo

import (
	"errors"
	"github.com/crushedpixel/ferry"
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
		db:          db,
		controllers: make(map[*Resource]*Controller),
		registry:    make(internal.SchemaRegistry),
		resources:   make(map[*internal.Schema]*Resource),
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

func (app *Application) Bridge(f *ferry.Ferry, namespace string) {
	// prepend slash to namespace
	if len(namespace) < 1 || namespace[0] != '/' {
		namespace = "/" + namespace
	}
	// remove trailing slash from namespace
	if namespace[len(namespace)-1] == '/' {
		namespace = namespace[:len(namespace)-1]
	}

	for resource, controller := range app.controllers {
		prefix := namespace + "/" + resource.JSONAPIName()

		if len(controller.indexAction) > 0 {
			f.GET(prefix, controller.indexAction.toFerry(app, controller))
		}
		if len(controller.showAction) > 0 {
			f.GET(prefix+"/{id}", controller.showAction.toFerry(app, controller))
		}
		if len(controller.createAction) > 0 {
			f.POST(prefix, controller.createAction.toFerry(app, controller))
		}
		if len(controller.updateAction) > 0 {
			f.PATCH(prefix+"/{id}", controller.updateAction.toFerry(app, controller))
		}
		if len(controller.deleteAction) > 0 {
			f.DELETE(prefix+"/{id}", controller.deleteAction.toFerry(app, controller))
		}

		for route, handlers := range controller.customActions {
			if len(handlers) > 0 {
				f.Handle(route.method, prefix+route.path, handlers.toFerry(app, controller))
			}
		}
	}
}
