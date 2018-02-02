package jargo

type ActionType int8

const (
	ActionTypeIndex ActionType = iota
	ActionTypeShow
	ActionTypeCreate
	ActionTypeUpdate
	ActionTypeDelete
)

type Request interface {
	// Resource returns the JSON API name
	// of the resource the client is accessing.
	Resource() string
	// ResourceId returns the requested resource id.
	// Only valid for Show, Update and Delete actions.
	ResourceId() int64
	// ActionType returns the action type
	// of the request.
	ActionType() ActionType

	// FieldSet returns the sparse fieldset requested
	// by the client.
	FieldSet() *FieldSet
	// Filters returns the filters requested by the client.
	// Only valid for Index actions.
	Filters() *Filters
	// Pagination returns the pagination settings
	// requested by the client.
	// Only valid for Index actions.
	Pagination() *Pagination
	// SortFields returns the sort settings
	// requested by the client.
	// Only valid for Index actions.
	SortFields() *SortFields

	// Payload returns the JSON API payload
	// sent by the client.
	// Only valid for Create and Update actions.
	Payload() string
}
