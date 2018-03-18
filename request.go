package jargo

import (
	"github.com/crushedpixel/ferry"
	"github.com/go-pg/pg"
)

type Request struct {
	*ferry.Request
	application *Application
	resource    *Resource
}

func (r *Request) Application() *Application {
	return r.application
}

func (r *Request) DB() *pg.DB {
	return r.application.db
}

func (r *Request) Resource() *Resource {
	return r.resource
}

type IndexRequest struct {
	*Request
	fields     *FieldSet
	filters    *Filters
	pagination Pagination
}

func (r *IndexRequest) Fields() *FieldSet {
	return r.fields
}

func (r *IndexRequest) Filters() *Filters {
	return r.filters
}

func (r *IndexRequest) Pagination() Pagination {
	return r.pagination
}

type ShowRequest struct {
	*Request
	fields     *FieldSet
	resourceId string
}

func (r *ShowRequest) Fields() *FieldSet {
	return r.fields
}

func (r *ShowRequest) ResourceId() string {
	return r.resourceId
}

type CreateRequest struct {
	*Request
	fields *FieldSet
}

func (r *CreateRequest) Fields() *FieldSet {
	return r.fields
}

type UpdateRequest struct {
	*Request
	fields     *FieldSet
	resourceId string
}

func (r *UpdateRequest) Fields() *FieldSet {
	return r.fields
}

func (r *UpdateRequest) ResourceId() string {
	return r.resourceId
}

type DeleteRequest struct {
	*Request
	resourceId string
}

func (r *DeleteRequest) ResourceId() string {
	return r.resourceId
}
