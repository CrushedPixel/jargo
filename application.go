package jargo

import (
	"crushedpixel.net/margo"
	"github.com/go-pg/pg"
	"errors"
	"log"
	"github.com/go-pg/pg/orm"
)

const (
	defaultPageSize = 25
)

type Application struct {
	*margo.Server

	ran bool

	DB          *pg.DB
	Controllers []*Controller
	PageSize    int
}

func NewApplication(db *pg.DB) *Application {
	server := margo.NewServer()
	return &Application{server, false, db, []*Controller{}, defaultPageSize}
}

func (app *Application) Run(addr ...string) error {
	if app.ran {
		return errors.New("application can't be run multiple times")
	}
	app.ran = true

	log.Println("starting jargo application")

	app.Use(injectApplicationMiddleware(app))

	for _, c := range app.Controllers {
		// create table for controller's model
		app.DB.CreateTable(c.Model, &orm.CreateTableOptions{IfNotExists: true})

		// register actions
		for _, e := range c.toEndpoints() {
			app.Register(e)
		}
	}

	return app.Server.Run(addr...)
}
