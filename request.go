package jargo

import (
	"io"
)

type ActionType int8

const (
	ActionTypeIndex ActionType = iota
	ActionTypeShow
	ActionTypeCreate
	ActionTypeUpdate
	ActionTypeDelete
)

type Request struct {
	// ActionType is the action type
	// of the request.
	ActionType ActionType

	// ResourceName is the JSON API name
	// of the resource the client is accessing.
	ResourceName string
	// ResourceId is the requested resource id.
	// Only valid for Show, Update and Delete actions.
	ResourceId string

	// FieldSet is a map of resource names
	// to field names according to the JSON API spec.
	// Only valid for Index, Show, Create and Update actions.
	// http://jsonapi.org/format/#fetching-sparse-fieldsets
	Fields map[string][]string
	// Filters is a map of filters requested by the client.
	// Only valid for Index actions.
	// http://jsonapi.org/format/#fetching-filtering
	Filters map[string]map[string][]string
	// Pagination are the pagination settings
	// requested by the client.
	// Only valid for Index actions.
	// http://jsonapi.org/format/#fetching-pagination
	Pagination map[string]string
	// Sort is a map of JSON API field names
	// to sort direction (true being ascending),
	// according to JSON API spec.
	// Only valid for Index actions.
	// http://jsonapi.org/format/#fetching-sorting
	Sort map[string]bool

	// Payload is the JSON API payload
	// sent by the client.
	// Only valid for Create and Update actions.
	Payload io.Reader

	// Header is a map of key-value pairs
	// sent with the request.
	Header map[string][]string
}
