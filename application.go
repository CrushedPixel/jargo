package jargo

import (
	"context"
	"errors"
	"fmt"
	"github.com/crushedpixel/jargo/internal"
	"github.com/go-pg/pg"
	"gopkg.in/go-playground/validator.v9"
	"reflect"
)

var (
	errNoResponse        = errors.New("the last HandlerFunc returned a nil Response")
	errAppAlreadyRunning = errors.New("application is already running")
	errAppNotRunning     = errors.New("application must be running to be able to handle http requests")
)

// Application is the central component of jargo.
// It contains all Controllers which are responsible
// for request handling.
type Application struct {
	db *pg.DB

	controllers map[*Resource]*Controller

	registry  internal.SchemaRegistry
	resources map[*internal.Schema]*Resource

	paginationStrategies *PaginationStrategies
	maxPageSize          int
	validate             *validator.Validate

	// running indicates whether the Application
	// is currently able to handle requests.
	running bool

	resourceExpirers *resourceExpirers
}

// NewApplication returns a new Application
// for the given Options.
func NewApplication(options Options) *Application {
	o := &options
	o.setDefaults()

	return &Application{
		controllers: make(map[*Resource]*Controller),
		registry:    make(internal.SchemaRegistry),
		resources:   make(map[*internal.Schema]*Resource),

		db: o.DB,

		paginationStrategies: o.PaginationStrategies,
		maxPageSize:          o.MaxPageSize,
		validate:             o.Validate,
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
// Returns an error if model is not an instance
// of a properly annotated Resource Model.
func (app *Application) RegisterResource(model interface{}) (*Resource, error) {
	s, err := app.registry.RegisterSchema(reflect.TypeOf(model))
	if err != nil {
		return nil, fmt.Errorf("error registering schema: %s", err.Error())
	}

	if resource, ok := app.resources[s]; ok {
		return resource, nil
	}

	for _, schema := range app.registry {
		if _, ok := app.resources[schema]; !ok {
			resource := &Resource{schema: schema}
			if err := resource.Initialize(app.DB()); err != nil {
				return nil, fmt.Errorf("error initializing resource %s: %s",
					schema.JSONAPIName(), err.Error())
			}

			// if app is already running,
			// create resource expirer for resource
			// if needed
			if app.running {
				if ef := resource.schema.ExpireField(); ef != nil {
					app.resourceExpirers.addExpirer(resource)
				}
			}

			app.resources[schema] = resource
		}
	}

	return app.resources[s], nil
}

// MustRegisterResource calls RegisterResource,
// panicking if it encounters an error.
func (app *Application) MustRegisterResource(model interface{}) *Resource {
	r, err := app.RegisterResource(model)
	if err != nil {
		panic(err)
	}
	return r
}

// Run handles incoming requests until ctx is done.
// This must be called before serving the Application.
// This is a blocking operation.
func (app *Application) Run(ctx context.Context) {
	if app.running {
		panic(errAppAlreadyRunning)
	}

	app.running = true
	app.resourceExpirers = newResourceExpirers(app, ctx)

	// start a resource expirer for all registered resources
	// that have an expire field
	for _, r := range app.resources {
		if ef := r.schema.ExpireField(); ef != nil {
			app.resourceExpirers.addExpirer(r)
		}
	}

	// wait until context is done
	<-ctx.Done()
}
