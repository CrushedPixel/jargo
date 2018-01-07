package jargo

import (
	"errors"
	"fmt"
	"github.com/crushedpixel/jargo/api"
	"github.com/crushedpixel/jargo/internal"
	"github.com/crushedpixel/margo"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg"
	"gopkg.in/go-playground/validator.v9"
	"log"
	"net/url"
	"reflect"
)

const defaultPageSize = 25

type Application struct {
	*margo.Application
	registry internal.Registry

	DB          *pg.DB
	Controllers []*Controller

	// maximum page size for pagination
	MaxPageSize int

	// Validate instance used by default create
	// and update handlers to validate user payloads
	Validate *validator.Validate

	ran bool
}

func NewApplication(db *pg.DB) *Application {
	server := margo.NewApplication()
	server.ErrorHandler = defaultErrorHandler
	return &Application{
		Application: server,
		registry:    make(internal.Registry),
		DB:          db,
		Controllers: []*Controller{},
		MaxPageSize: defaultPageSize,
		Validate:    validator.New(),
	}
}

// registers a resource with the application, creating the respective
// database tables if they don't exist yet.
func (app *Application) RegisterResource(model interface{}) (resource api.Resource, err error) {
	resource, err = app.registry.RegisterResource(reflect.TypeOf(model))

	err = resource.CreateTable(app.DB)
	if err != nil {
		resource = nil
	}
	return
}

func (app *Application) AddController(c *Controller) {
	app.Controllers = append(app.Controllers, c)
}

func (app *Application) Run(addr ...string) error {
	if app.ran {
		return errors.New("application can't be run multiple times")
	}
	app.ran = true
	log.Println("starting jargo application")
	app.Use(injectApplicationMiddleware(app))

	for _, c := range app.Controllers {
		c.initialize(app)
	}

	return app.Application.Run(addr...)
}

func (app *Application) ParsePagination(query url.Values) (api.Pagination, error) {
	return internal.ParsePagination(query, app.MaxPageSize)
}

func defaultErrorHandler(c *gin.Context, r interface{}) {
	var err error
	var ok bool

	err, ok = r.(error)
	if !ok {
		log.Println(fmt.Sprintf("%s", r))
		err = api.ErrInternalServerError
	}

	res := api.NewErrorResponse(err)
	err = res.Send(c)
	if err != nil {
		panic(err)
	}
}
