package api

import (
	"github.com/go-pg/pg"
	"crushedpixel.net/jargo/internal"
	"crushedpixel.net/margo"
	"net/url"
)

type Resource interface {
	Name() string

	// parses a FieldSet for this Resource from fields query parameters
	ParseFieldSet(query url.Values) (FieldSet, error)
	// parses SortFields for this Resource from sort query parameters
	ParseSortFields(query url.Values) (SortFields, error)
	// parses Filters for this Resource from filter query parameters
	ParseFilters(query url.Values) (Filters, error)

	CreateTable(*pg.DB) error

	Select(*pg.DB) Query
	SelectOne(*pg.DB) Query
	SelectById(*pg.DB, interface{}) Query

	Response(data interface{}, fields FieldSet) margo.Response
	ResponseAllFields(data interface{}) margo.Response
	ResponseWithStatusCode(data interface{}, fields FieldSet, status int) margo.Response
}

var _ Resource = (*internal.Resource)(nil)
