package jargo

import (
	"errors"
	"github.com/crushedpixel/jargo/internal"
	"github.com/go-pg/pg"
	"gopkg.in/go-playground/validator.v9"
	"reflect"
)

var errNoResponse = errors.New("the last HandlerFunc returned a nil Response")

// Application is the central component of jargo.
type Application struct {
	db *pg.DB

	controllers map[*Resource]*Controller

	registry  internal.SchemaRegistry
	resources map[*internal.Schema]*Resource

	paginationStrategies *PaginationStrategies
	maxPageSize          int
	validate             *validator.Validate
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
