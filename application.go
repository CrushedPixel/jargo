package jargo

import (
	"context"
	"errors"
	"github.com/crushedpixel/jargo/internal"
	"github.com/go-pg/pg"
	"gopkg.in/go-playground/validator.v9"
	"reflect"
)

var (
	errNoResponse        = errors.New("the last HandlerFunc returned a nil Response")
	errAppAlreadyRunning = errors.New("realtime instance is already running")
	errAppNotRunning     = errors.New("realtime instance must be started to be able to handle http requests")
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

	// ctx is the Context bound to the Application's goroutines.
	ctx context.Context
	// cancel is the CancelFunc to call to release
	// the Application's goroutines.
	cancel context.CancelFunc

	// running indicates whether the Application
	// is currently able to handle requests.
	running bool
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

			// if app is already running,
			// create resource expirer for resource
			// if needed
			if app.running {
				if ef := resource.schema.ExpireField(); ef != nil {
					newResourceExpirer(app.ctx, app, resource)
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

// Start prepares the Application to handle incoming requests.
// This must be called before serving the Application.
//
// After handling is done, Release should be called to stop all internal
// goroutines.
func (app *Application) Start() {
	if app.running {
		panic(errAppAlreadyRunning)
	}

	app.running = true

	app.ctx, app.cancel = context.WithCancel(context.Background())
	// start a resource expirer for all registered resources
	// that have an expire field
	for _, r := range app.resources {
		if ef := r.schema.ExpireField(); ef != nil {
			newResourceExpirer(app.ctx, app, r)
		}
	}
}

// Release stops all internal goroutines.
// Should be called after serving is done.
func (app *Application) Release() {
	if !app.running {
		panic(errAppNotRunning)
	}
	app.cancel()
	app.running = false
}
