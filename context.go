package jargo

import "github.com/go-pg/pg"

type Context struct {
	application *Application
	resource    *Resource
	data        map[string]interface{}
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

func (c *Context) Get(key string) interface{} {
	return c.data[key]
}

func (c *Context) Set(key string, value interface{}) {
	c.data[key] = value
}
