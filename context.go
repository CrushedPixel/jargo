package jargo

import (
	"github.com/crushedpixel/ferry"
	"github.com/go-pg/pg"
)

type Context struct {
	*ferry.Context

	application *Application
	resource    *Resource
}

func (c *Context) Application() *Application {
	return c.application
}

func (c *Context) DB() *pg.DB {
	return c.application.db
}

func (c *Context) Resource() *Resource {
	return c.resource
}
