package jargo

import (
	"io"
)

type IndexRequest struct {
	Fields     *FieldSet
	Filters    *Filters
	SortFields *SortFields
	Pagination *Pagination
}

type ShowRequest struct {
	Fields     *FieldSet
	ResourceId int64
}

type CreateRequest struct {
	Fields  *FieldSet
	Payload io.Reader
}

type UpdateRequest struct {
	Fields     *FieldSet
	ResourceId int64
	Payload    io.Reader
}

type DeleteRequest struct {
	ResourceId int64
}
