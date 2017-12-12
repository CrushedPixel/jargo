package jargo

import (
	"crushedpixel.net/margo"
	"github.com/go-pg/pg"
	"errors"
	"log"
)

const (
	defaultPageSize = 25
)

type Application struct {
	*margo.Server

	DB          *pg.DB
	Controllers []*Controller
	MaxPageSize int

	ran bool
}

func NewApplication(db *pg.DB) *Application {
	server := margo.NewServer()
	return &Application{server, db, []*Controller{}, defaultPageSize, false}
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
