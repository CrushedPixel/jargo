package jargo

import (
	"crushedpixel.net/margo"
	"github.com/go-pg/pg"
	"errors"
	"log"
	"github.com/gin-gonic/gin"
	"fmt"
	"crushedpixel.net/jargo/internal/resource"
	"crushedpixel.net/jargo/api"
	"crushedpixel.net/jargo/internal"
)

const (
	defaultPageSize = 25
)

type Application struct {
	*margo.Server
	registry *internal.Registry

	DB          *pg.DB
	Controllers []*Controller
	MaxPageSize int

	ran bool
}

var defaultErrorHandler margo.ErrorHandlerFunc = func(c *gin.Context, r interface{}) {
	var err error
	var ok bool

	err, ok = r.(error)
	if !ok {
		println(fmt.Sprintf("%s", r)) // TODO: proper logging
		err = ApiErrInternalServerError
	}

	res := NewErrorResponse(err)
	err = res.Send(c)
	if err != nil {
		panic(err)
	}
}

func NewApplication(db *pg.DB) *Application {
	server := margo.NewServer()
	server.ErrorHandler = defaultErrorHandler
	registry := resource.NewRegistry()
	return &Application{
		Server:      server,
		registry:    registry,
		DB:          db,
		Controllers: []*Controller{},
		MaxPageSize: defaultPageSize,
	}
}

func (app *Application) RegisterResource(model interface{}) (api.Resource, error) {
	return app.registry.RegisterResource(model)
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
