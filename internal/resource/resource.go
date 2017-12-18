package resource

import "reflect"

type Resource struct {
	// TODO
}

type fieldType int

const (
	_         fieldType = iota
	id
	attribute
	has
	belongsTo
	many2many
)

type field struct {
	structField *reflect.StructField

	typ fieldType

	name   string // jsonapi attribute name. typ=attribute,relationship only
	column string // table column name.      typ=attribute only

	readonly bool
	sort     bool
	filter   bool

	sqlNotnull bool   // typ=attribute,relationship only
	sqlUnique  bool   // typ=attribute,relationship only
	sqlDefault string // typ=attribute only

	pgFk        string // typ=has only
	pgJoinTable string // typ=many2many only
}
