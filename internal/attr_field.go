package internal

import (
	"reflect"
	"errors"
	"fmt"
	"crushedpixel.net/jargo/internal/parser"
	"github.com/iancoleman/strcase"
)

var (
	errJsonapiOptionOnUnexportedField = errors.New("jsonapi-related option on unexported field")
	errInvalidColumnName              = errors.New("column name may only consist of [0-9,a-z,A-Z$_]")
)

func errInvalidAttrFieldType(p reflect.Type) error {
	return errors.New(fmt.Sprintf("invalid type for attribute field: %s", p))
}

type attrField struct {
	*baseField

	column     string // sql column name
	sqlDefault string
}

func newAttrField(schema *schema, f *reflect.StructField) (field, error) {
	base, err := newBaseField(schema, f)
	if err != nil {
		return nil, err
	}

	column := base.name
	if !base.jsonapiExported {
		column = strcase.ToSnake(base.fieldName)
	}
	field := &attrField{
		baseField: base,
		column:    column,
	}

	// validate field type
	typ := f.Type
	if !isValidAttrType(typ) {
		return nil, errInvalidAttrFieldType(typ)
	}

	parsed := parser.ParseJargoTag(f.Tag.Get(jargoFieldTag))

	// parse options
	for option, value := range parsed.Options {
		switch option {
		case optionColumn:
			field.column = value
		case optionDefault:
			field.sqlDefault = value
		case optionReadonly, optionSort, optionFilter,
			optionOmitempty, optionNotnull, optionUnique:
			// these were handled when parsing the baseField
			// and should therefore not trigger the default handler.
		default:
			return nil, errDisallowedOption(option)
		}
	}

	// validate sql column
	if !isValidSQLName(field.column) {
		return nil, errInvalidColumnName
	}

	// finally, generate jsonapi and pg attribute fields
	field.jsonapiF = jsonapiAttrFields(field)
	field.pgF = pgAttrFields(field)

	return field, nil
}

func (f *attrField) pgColumn() string {
	return fmt.Sprintf("%s.%s", f.schema.alias, f.column)
}

func (f *attrField) jsonapiJoinFields() ([]reflect.StructField, error) {
	// the jsonapi join model only needs the id field
	return []reflect.StructField{}, nil
}

func jsonapiAttrFields(f *attrField) []reflect.StructField {
	if f.name == unexportedFieldName {
		return []reflect.StructField{}
	}

	tag := fmt.Sprintf(`jsonapi:"attr,%s`, f.name)
	if f.jsonapiOmitempty {
		tag += `,omitempty`
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
	if f.sqlNotnull {
		tag += ",notnull"
	}
	if f.sqlUnique {
		tag += ",unique"
	}
	if f.sqlDefault != "" {
		tag += fmt.Sprintf(",default:%s", f.sqlDefault)
	}
	tag += `"`

	field := reflect.StructField{
		Name: f.fieldName,
		Type: f.fieldType,
		Tag:  reflect.StructTag(tag),
	}

	return []reflect.StructField{field}
}

func (f *attrField) createInstance() fieldInstance {
	return &attrFieldInstance{
		field: f,
	}
}

func isValidAttrType(typ reflect.Type) bool {
	switch reflect.New(typ).Elem().Interface().(type) {
	case bool, int, int8, int16, int32, int64, uint, uint8,
	uint16, uint32, uint64, float32, float64, string:
		return true
	default:
		return false
	}
}

type attrFieldInstance struct {
	field *attrField
	value interface{}
}

func (i *attrFieldInstance) parseResourceModel(instance *resourceModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	return i.parse(instance.value)
}

func (i *attrFieldInstance) applyToResourceModel(instance *resourceModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	return i.apply(instance.value)
}

func (i *attrFieldInstance) parseJoinResourceModel(instance *resourceModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	return i.parse(instance.value)
}

func (i *attrFieldInstance) applyToJoinResourceModel(instance *resourceModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	return i.apply(instance.value)
}

func (i *attrFieldInstance) parseJsonapiModel(instance *jsonapiModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	if !i.field.jsonapiExported {
		return nil
	}
	return i.parse(instance.value)
}

func (i *attrFieldInstance) applyToJsonapiModel(instance *jsonapiModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	if !i.field.jsonapiExported {
		return nil
	}
	return i.apply(instance.value)
}

func (i *attrFieldInstance) parseJoinJsonapiModel(instance *joinJsonapiModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	// only the id field exists in join jsonapi models
	return nil
}

func (i *attrFieldInstance) applyToJoinJsonapiModel(instance *joinJsonapiModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	// only the id field exists in join jsonapi models
	return nil
}

func (i *attrFieldInstance) parsePGModel(instance *pgModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	return i.parse(instance.value)
}

func (i *attrFieldInstance) applyToPGModel(instance *pgModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	return i.apply(instance.value)
}

func (i *attrFieldInstance) parseJoinPGModel(instance *joinPGModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	return i.parse(instance.value)
}

func (i *attrFieldInstance) applyToJoinPGModel(instance *joinPGModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	return i.apply(instance.value)
}

func (i *attrFieldInstance) parse(v *reflect.Value) error {
	if v.IsNil() {
		return nil
	}
	i.value = v.Elem().FieldByName(i.field.fieldName).Interface()
	return nil
}

func (i *attrFieldInstance) apply(v *reflect.Value) error {
	if v.IsNil() {
		panic(errors.New("struct pointer must not be nil"))
	}
	v.Elem().FieldByName(i.field.fieldName).Set(reflect.ValueOf(i.value))
	return nil
}
