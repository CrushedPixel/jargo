package internal

import (
	"errors"
	"fmt"
	"github.com/c9s/inflect"
	"github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"
	"reflect"
	"time"
)

const validationTag = "validate"

var (
	errJsonapiOptionOnUnexportedField = errors.New("jsonapi-related option on unexported field")
	errInvalidColumnName              = errors.New("column name may only consist of [0-9,a-z,A-Z$_]")
	errNonNullableTypeDefault         = errors.New(`"default" option may only be used on pointer types`)
	errNotnullWithoutDefault          = errors.New(`"notnull" option may only be used in conjunction with the "default" option. use a primitive type instead`)

	errCreatedAtDefaultForbidden   = errors.New(`"default" option may not be used in conjunction with "createdAt""`)
	errUpdatedAtDefaultForbidden   = errors.New(`"default" option may not be used in conjunction with "updatedAt""`)
	errCreatedAtUpdatedAtExclusive = errors.New(`"createdAt" and "updatedAt" options are mutually exclusive`)
	errCreatedAtUpdatedAtType      = errors.New(`"createdAt" and "updatedAt" options are only allowed on fields of type *time.Time`)
	errCreatedAtUpdatedAtWritable  = errors.New(`"createdAt" and "updatedAt" options are only allowed on writable (non-readonly) fields`)

	autoTimestampsType = reflect.TypeOf(&time.Time{})
)

type attrField struct {
	*baseField

	column     string // sql column name
	sqlDefault string

	// whether the field should have a NOT NULL constraint
	// although it is a pointer type.
	// this may be used to use DEFAULT values on NOT NULL fields
	// by inserting nil pointer values.
	notnull bool

	validation string
}

func newAttrField(schema *Schema, f *reflect.StructField) SchemaField {
	base := newBaseField(schema, f)

	// determine default column name.
	// defaults to underscored jsonapi member name.
	// if jsonapi member name is "-" (unexported field),
	// defaults to underscored struct field name.
	var column string
	if base.jsonapiExported {
		column = inflect.Underscore(base.name)
	} else {
		column = inflect.Underscore(base.fieldName)
	}

	field := &attrField{
		baseField: base,
		column:    column,
	}

	parsed := parseJargoTag(f.Tag.Get(jargoFieldTag))

	var createdAt, updatedAt bool

	// parse default option before others,
	// so it is set before notnull is processed
	if value, ok := parsed.Options[optionDefault]; ok {
		if !isNullable(field.fieldType) {
			// a default value may only be set for
			// pointer types, to avoid zero values
			// being omitted by go-pg (go-pg#790)
			panic(errNonNullableTypeDefault)
		}
		field.sqlDefault = value
	}

	// parse options
	for option, value := range parsed.Options {
		switch option {
		case optionColumn:
			field.column = value
		case optionNotnull:
			field.notnull = parseBoolOption(value)
			if field.notnull && field.sqlDefault == "" {
				panic(errNotnullWithoutDefault)
			}
		case optionCreatedAt:
			createdAt = parseBoolOption(value)
		case optionUpdatedAt:
			updatedAt = parseBoolOption(value)
		case optionReadonly, optionNoSort, optionNoFilter,
			optionOmitempty, optionUnique, optionDefault:
			// these were handled when parsing the baseField
			// and should therefore not trigger the default handler.
		default:
			panic(errDisallowedOption(option))
		}
	}

	// validate createdAt and updatedAt tags
	if createdAt && updatedAt {
		panic(errCreatedAtUpdatedAtExclusive)
	}
	if field.sqlDefault != "" && createdAt {
		panic(errCreatedAtDefaultForbidden)
	}
	if field.sqlDefault != "" && updatedAt {
		panic(errUpdatedAtDefaultForbidden)
	}

	if createdAt || updatedAt {
		field.notnull = true

		if field.fieldType != autoTimestampsType {
			panic(errCreatedAtUpdatedAtType)
		}

		// disallow explicit writable (readonly:false) option
		if _, ok := parsed.Options[optionReadonly]; ok && field.jargoWritable {
			panic(errCreatedAtUpdatedAtWritable)
		}
		// auto timestamps may not be changed by api users
		field.jargoWritable = false

		// set default to "NOW()" for createdAt and updatedAt columns
		field.sqlDefault = "NOW()"
	}

	// validate sql column
	if !isValidSQLName(field.column) {
		panic(errInvalidColumnName)
	}

	// store "validate" struct tag
	field.validation = f.Tag.Get(validationTag)

	// finally, generate jsonapi and pg attribute fields
	field.jsonapiF = jsonapiAttrFields(field)
	field.pgF = pgAttrFields(field)

	// wrap updatedAt fields in updatedAtField struct
	// to implement afterCreateTable
	if updatedAt {
		return &updatedAtField{
			field,
		}
	}

	return field
}

func (f *attrField) ColumnName() string {
	return f.column
}

func (f *attrField) PGSelectColumn() string {
	return fmt.Sprintf("%s.%s", f.schema.alias, f.column)
}

func (f *attrField) PGFilterColumn() string {
	return f.PGSelectColumn()
}

func (f *attrField) jsonapiJoinFields() []reflect.StructField {
	// the jsonapi join model only needs the id field
	return []reflect.StructField{}
}

func (f *attrField) Sortable() bool {
	// override sortable to take f.notnull flag into account
	return f.jargoSortable && !f.isNullable()
}

func (f *attrField) isNullable() bool {
	return isNullable(f.fieldType) && !f.notnull
}

func jsonapiAttrFields(f *attrField) []reflect.StructField {
	if f.name == unexportedFieldName {
		return []reflect.StructField{}
	}

	tag := fmt.Sprintf(`jsonapi:"attr,%s`, f.name)
	if f.jsonapiOmitempty {
		tag += `,omitempty`
	}
	if isTimeField(f.fieldType) {
		// the jsonapi spec recommends using ISO8601 for date/time formatting.
		// see http://jsonapi.org/recommendations/#date-and-time-fields
		tag += `,iso8601`
	}
	tag += `"`

	field := reflect.StructField{
		Name: f.fieldName,
		Type: f.fieldType,
		Tag:  reflect.StructTag(tag),
	}

	return []reflect.StructField{field}
}

func pgAttrFields(f *attrField) []reflect.StructField {
	tag := fmt.Sprintf(`sql:"%s`, f.column)
	if !f.isNullable() {
		tag += ",notnull"
	}
	if f.sqlUnique {
		tag += ",unique"
	}
	if f.sqlDefault != "" {
		tag += fmt.Sprintf(",default:%s", f.sqlDefault)
	}
	if isUUIDField(f.fieldType) {
		tag += ",type:uuid"
	}
	tag += `"`

	field := reflect.StructField{
		Name: f.fieldName,
		Type: f.fieldType,
		Tag:  reflect.StructTag(tag),
	}

	return []reflect.StructField{field}
}

func (f *attrField) createInstance() schemaFieldInstance {
	return &attrFieldInstance{
		field: f,
	}
}

// isNullable returns whether typ represents a nullable type.
func isNullable(typ reflect.Type) bool {
	return typ.Kind() == reflect.Ptr
}

// isTimeField returns whether typ is a time.Time field.
func isTimeField(typ reflect.Type) bool {
	// pointer types are allowed
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	switch reflect.New(typ).Elem().Interface().(type) {
	case time.Time:
		return true
	default:
		return false
	}
}

// isUUIDField returns whether typ is a uuid.UUID field.
func isUUIDField(typ reflect.Type) bool {
	// pointer types are allowed
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	switch reflect.New(typ).Elem().Interface().(type) {
	case uuid.UUID:
		return true
	default:
		return false
	}
}

type attrFieldInstance struct {
	field *attrField
	value interface{}
}

func (i *attrFieldInstance) parentField() SchemaField {
	return i.field
}

func (i *attrFieldInstance) sortValue() interface{} {
	return i.value
}

func (i *attrFieldInstance) parseResourceModel(instance *resourceModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.parse(instance.value)
}

func (i *attrFieldInstance) applyToResourceModel(instance *resourceModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.apply(instance.value)
}

func (i *attrFieldInstance) parseJoinResourceModel(instance *resourceModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.parse(instance.value)
}

func (i *attrFieldInstance) applyToJoinResourceModel(instance *resourceModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.apply(instance.value)
}

func (i *attrFieldInstance) parseJsonapiModel(instance *jsonapiModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	if i.field.jsonapiExported {
		i.parse(instance.value)
	}
}

func (i *attrFieldInstance) applyToJsonapiModel(instance *jsonapiModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	if i.field.jsonapiExported {
		i.apply(instance.value)
	}
}

func (i *attrFieldInstance) parseJoinJsonapiModel(instance *joinJsonapiModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	// only the id field exists in join jsonapi models
}

func (i *attrFieldInstance) applyToJoinJsonapiModel(instance *joinJsonapiModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	// only the id field exists in join jsonapi models
}

func (i *attrFieldInstance) parsePGModel(instance *pgModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.parse(instance.value)
}

func (i *attrFieldInstance) applyToPGModel(instance *pgModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.apply(instance.value)
}

func (i *attrFieldInstance) parseJoinPGModel(instance *joinPGModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.parse(instance.value)
}

func (i *attrFieldInstance) applyToJoinPGModel(instance *joinPGModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.apply(instance.value)
}

func (i *attrFieldInstance) parse(v *reflect.Value) {
	if !v.IsNil() {
		i.value = v.Elem().FieldByName(i.field.fieldName).Interface()
	}
}

func (i *attrFieldInstance) apply(v *reflect.Value) {
	if v.IsNil() {
		panic(errNilPointer)
	}
	if i.value != nil {
		v.Elem().FieldByName(i.field.fieldName).Set(reflect.ValueOf(i.value))
	}
}

func (i *attrFieldInstance) validate(validate *validator.Validate) error {
	return validate.Var(i.value, i.field.validation)
}
