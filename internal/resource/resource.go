package resource

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
	typ fieldType

	name   string // jsonapi attribute name
	column string // table column name

	readonly bool
	sort     bool
	filter   bool

	notnull bool
	unique  bool
	defavlt string

	fk        string // has, belongsTo only
	joinTable string // many2many only
}
