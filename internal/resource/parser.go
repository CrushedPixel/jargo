package resource

import (
	"errors"
	"reflect"
	"crushedpixel.net/jargo/internal/parser"
	"github.com/iancoleman/strcase"
	"strconv"
	"fmt"
)

const (
	primaryFieldName    = "Id"
	unexportedFieldName = "-"

	jargoFieldTag   = "jargo"
	optionTable     = "table"
	optionAlias     = "alias"
	optionColumn    = "column"
	optionHas       = "has"
	optionBelongsTo = "belongsTo"
	optionMany2Many = "many2many"
	optionReadonly  = "readonly"
	optionSort      = "sort"
	optionFilter    = "filter"
	optionOmitempty = "omitempty"
	optionNotnull   = "notnull"
	optionUnique    = "unique"
	optionDefault   = "default"
)

type resourceDefinition struct {
	fields []*fieldDefinition
	name   string
	table  string
	alias  string
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

type fieldDefinition struct {
	structField *reflect.StructField

	typ fieldType

	name   string // jsonapi attribute name. typ=attribute,relationship only
	column string // table column name.      typ=attribute,belongsTo only

	readonly bool
	sort     bool
	filter   bool

	jsonOmitempty bool // typ=attribute,relationship only

	sqlNotnull bool   // typ=attribute,relationship only
	sqlUnique  bool   // typ=attribute,relationship only
	sqlDefault string // typ=attribute only

	pgFk        string // typ=has only
	pgJoinTable string // typ=many2many only
}

var errInvalidModelType = errors.New("model has to be struct")
var errMissingIdField = errors.New("missing id field")
var errInvalidIdType = errors.New("id field must be of type int64")
var errUnannotatedIdField = errors.New("id field is missing jargo annotation")
var errInvalidMemberName = errors.New(`member name has to adhere to the jsonapi specification and not include characters marked as "not recommended"`)
var errInvalidTableName = errors.New("table name may only consist of [0-9,a-z,A-Z$_]")
var errInvalidTableAlias = errors.New("alias may only consist of [0-9,a-z,A-Z$_]")
var errAliasEqualsTableName = errors.New("alias may not be equal to table name")
var errInvalidColumnName = errors.New("column name may only consist of [0-9,a-z,A-Z$_]")
var errStructType = errors.New("structs are not allowed as a field type. for a relation, use a struct pointer instead")
var errMissingRelationTag = errors.New("missing relation tag on struct pointer or slice field. has, belongsTo or many2many option is required")
var errInvalidHasType = errors.New("has relation field has to be a struct pointer or slice of struct pointers")
var errInvalidBelongsToType = errors.New("belongsTo field has to be a struct pointer")
var errInvalidMany2ManyType = errors.New("many2many field has to be a slice of struct pointers")
var errMissingMany2ManyJoinTable = errors.New("missing many2many join table definition")
var errMultipleRelations = errors.New("multiple relation tags are not allowed")

func errDisallowedOption(option string) error {
	return errors.New(fmt.Sprintf(`option "%s" is not allowed here`, option))
}

func parseResourceStruct(model interface{}) (*resourceDefinition, error) {
	return parseResourceType(reflect.TypeOf(model))
}

func parseResourceType(t reflect.Type) (*resourceDefinition, error) {
	if t.Kind() != reflect.Struct {
		return nil, errInvalidModelType
	}

	// parse Id field
	rd, err := parseIdField(t)
	if err != nil {
		return nil, err
	}

	rd.fields = make([]*fieldDefinition, 0)

	// parse all attributes
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		field, err := parseAttrField(&f)
		if err != nil {
			return nil, err
		}

		rd.fields = append(rd.fields, field)
	}

	return rd, nil
}

func parseIdField(t reflect.Type) (*resourceDefinition, error) {
	idField, ok := t.FieldByName(primaryFieldName)
	if !ok {
		return nil, errMissingIdField
	}

	if idField.Type != reflect.TypeOf(int64(0)) {
		return nil, errInvalidIdType
	}

	idTag, ok := idField.Tag.Lookup(jargoFieldTag)
	if !ok {
		return nil, errUnannotatedIdField
	}

	singleName := strcase.ToSnake(t.Name())
	defaultName := pluralize(singleName)
	parsed := parser.ParseJargoTagDefaultName(idTag, defaultName)

	rd := &resourceDefinition{
		name:  parsed.Name,
		table: parsed.Name,
		alias: singleName,
	}

	for option, value := range parsed.Options {
		switch option {
		case optionTable:
			rd.table = value
		case optionAlias:
			rd.alias = value
		default:
			return nil, errDisallowedOption(option)
		}
	}

	// validate member name
	if !isValidJsonapiMemberName(rd.name) {
		return nil, errInvalidMemberName
	}

	// validate table name
	if !isValidSQLName(rd.table) {
		return nil, errInvalidTableName
	}

	// validate alias
	if !isValidSQLName(rd.alias) {
		return nil, errInvalidTableAlias
	}

	// ensure alias is not the same as table name
	if rd.alias == rd.table {
		return nil, errAliasEqualsTableName
	}

	return rd, nil
}

func parseAttrField(f *reflect.StructField) (*fieldDefinition, error) {
	if f.Name == primaryFieldName {
		return &fieldDefinition{
			structField: f,
			typ:         id,
		}, nil
	}

	defaultName := strcase.ToSnake(f.Name)
	parsed := parser.ParseJargoTagDefaultName(f.Tag.Get(jargoFieldTag), defaultName)

	field := &fieldDefinition{
		structField: f,
		name:        parsed.Name,

		readonly:      false,
		sort:          true,
		filter:        true,
		jsonOmitempty: false,
		sqlNotnull:    false,
		sqlUnique:     false,
		sqlDefault:    "",
		pgFk:          "",
		pgJoinTable:   "",
	}

	// get field type
	if fk, ok := parsed.Options[optionHas]; ok {
		field.typ = has
		field.pgFk = fk
	}

	if _, bt := parsed.Options[optionBelongsTo]; bt {
		if field.typ != 0 {
			return nil, errMultipleRelations
		}
		field.typ = belongsTo
	}

	if joinTable, ok := parsed.Options[optionMany2Many]; ok {
		if field.typ != 0 {
			return nil, errMultipleRelations
		}

		if joinTable == "" {
			return nil, errMissingMany2ManyJoinTable
		}

		if !isValidSQLName(joinTable) {
			return nil, errInvalidTableName
		}

		field.typ = many2many
		field.pgJoinTable = joinTable
	}

	// if field type is neither the id field nor a relation, it's an attribute
	if field.typ == 0 {
		field.typ = attribute

		// set default column name for attribute
		if parsed.Name == unexportedFieldName {
			field.column = defaultName
		} else {
			field.column = parsed.Name
		}
	}

	// validate value type for field type
	switch field.typ {
	case attribute:
		// all primitives except for pointers,
		// slices and structs are allowed
		switch f.Type.Kind() {
		case reflect.Struct:
			return nil, errStructType
		case reflect.Ptr, reflect.Slice:
			return nil, errMissingRelationTag
		}
	case has:
		// only struct pointers and
		// slices of struct pointers are allowed
		switch f.Type.Kind() {
		case reflect.Slice:
			if f.Type.Elem().Kind() != reflect.Ptr || f.Type.Elem().Elem().Kind() != reflect.Struct {
				return nil, errInvalidHasType
			}
		case reflect.Ptr:
			if f.Type.Elem().Kind() != reflect.Struct {
				return nil, errInvalidHasType
			}
		default:
			return nil, errInvalidHasType
		}
	case belongsTo:
		// only struct pointers are allowed
		if f.Type.Kind() != reflect.Ptr || f.Type.Elem().Kind() != reflect.Struct {
			return nil, errInvalidBelongsToType
		}
	case many2many:
		// only slices of struct pointers are allowed
		if f.Type.Kind() != reflect.Slice ||
			f.Type.Elem().Kind() != reflect.Ptr ||
			f.Type.Elem().Elem().Kind() != reflect.Struct {
			return nil, errInvalidMany2ManyType
		}
	}

	// parse rest of options
	for option, value := range parsed.Options {
		switch option {
		case optionColumn:
			if field.typ != attribute {
				return nil, errDisallowedOption(option)
			}
			field.column = value
		case optionReadonly:
			b, err := parseBool(value)
			if err != nil {
				return nil, err
			}
			field.readonly = b
		case optionSort:
			b, err := parseBool(value)
			if err != nil {
				return nil, err
			}
			field.sort = b
		case optionFilter:
			b, err := parseBool(value)
			if err != nil {
				return nil, err
			}
			field.filter = b
		case optionOmitempty:
			b, err := parseBool(value)
			if err != nil {
				return nil, err
			}
			field.jsonOmitempty = b
		case optionNotnull:
			b, err := parseBool(value)
			if err != nil {
				return nil, err
			}
			field.sqlNotnull = b
		case optionUnique:
			b, err := parseBool(value)
			if err != nil {
				return nil, err
			}
			field.sqlUnique = b
		case optionDefault:
			if field.typ != attribute {
				return nil, errDisallowedOption(option)
			}
			field.sqlDefault = value
		case optionBelongsTo, optionHas, optionMany2Many:
			// these were already handled when parsing the type
			// and do not trigger the default handler
		default:
			return nil, errDisallowedOption(option)
		}
	}

	// validate jsonapi member name
	if field.name != unexportedFieldName && !isValidJsonapiMemberName(field.name) {
		return nil, errInvalidMemberName
	}

	// validate column name
	if field.typ == attribute && !isValidSQLName(field.column) {
		return nil, errInvalidColumnName
	}

	return field, nil
}

func parseBool(val string) (bool, error) {
	if val == "" {
		return true, nil
	}

	return strconv.ParseBool(val)
}
