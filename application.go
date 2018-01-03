package jargo

import (
	"crushedpixel.net/margo"
	"github.com/go-pg/pg"
	"errors"
	"log"
	"github.com/gin-gonic/gin"
	"fmt"
	"crushedpixel.net/jargo/api"
	"crushedpixel.net/jargo/internal"
	"net/url"
	"reflect"
)

const (
	defaultPageSize = 25
)

type Application struct {
	*margo.Server
	registry internal.Registry

	DB          *pg.DB
	Controllers []*Controller
	MaxPageSize int

	ran bool
}

func defaultErrorHandler(c *gin.Context, r interface{}) {
	var err error
	var ok bool

	err, ok = r.(error)
	if !ok {
		println(fmt.Sprintf("%s", r)) // TODO: proper logging
		err = api.ErrInternalServerError
	}

	res := api.NewErrorResponse(err)
	err = res.Send(c)
	if err != nil {
		panic(err)
	}
}

func NewApplication(db *pg.DB) *Application {
	server := margo.NewServer()
	server.ErrorHandler = defaultErrorHandler
	registry := make(internal.Registry)
	return &Application{
		Server:      server,
		registry:    registry,
		DB:          db,
		Controllers: []*Controller{},
		MaxPageSize: defaultPageSize,
	}
}

func (app *Application) RegisterResource(model interface{}) (resource api.Resource, err error) {
	// internally, jargo panics when parsing an invalid resource.
	// to be a little bit more gracious to the user, we recover from those
	// and return them as an error value.
	defer func() {
		if r := recover(); r != nil {
			resource = nil

			switch x := r.(type) {
			case error:
				err = x
			default:
				panic(r)
			}
		}
	}()

	resource = app.registry.RegisterResource(reflect.TypeOf(model))
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

	return app.Server.Run(addr...)
}

func (app *Application) ParsePagination(query url.Values) (api.Pagination, error) {
	return internal.ParsePagination(query, app.MaxPageSize)
}
