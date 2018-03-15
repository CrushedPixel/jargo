package internal

import (
	"encoding"
	"errors"
	"fmt"
	"gopkg.in/go-playground/validator.v9"
	"reflect"
)

var (
	errInvalidIdType     = errors.New("id field type must be string or non-float numeric, or implement encoding.TextMarshaler and encoding.TextUnmarshaler")
	errNilPointer        = errors.New("resource must not be nil")
	errMismatchingSchema = errors.New("mismatching schema")
	emptyStructType      = reflect.TypeOf(new(struct{})).Elem()
)

const (
	IdFieldColumn      = "id"
	IdFieldJsonapiName = "id"
)

// implements field
type idField struct {
	schema *Schema

	fieldType reflect.Type
	// whether the field needs to be marshalled
	// into a string for jsonapi
	marshalling bool

	jsonapiF []reflect.StructField
	pgF      []reflect.StructField
}

func newIdField(schema *Schema, f *reflect.StructField) SchemaField {
	idf := &idField{
		schema:    schema,
		fieldType: f.Type,
	}

	valid, marshalling := isValidIdField(idf.fieldType)
	if !valid {
		panic(errInvalidIdType)
	}
	idf.marshalling = marshalling

	// generate jsonapi and pg attribute fields
	idf.jsonapiF = idf.jsonapiIdFields()
	idf.pgF = idf.pgIdFields()

	// wrap id fields with uuid type in uuidIdField
	// for afterCreateTable hook
	if isUUIDField(idf.fieldType) {
		return &uuidIdField{idf}
	}

	return idf
}

// isValidIdField returns whether typ is a valid type for an id field
// and whether it is a type that needs to be converted to string.
func isValidIdField(typ reflect.Type) (bool, bool) {
	// allow types allowed by jsonapi
	switch reflect.New(typ).Elem().Interface().(type) {
	case string, int, int8, int16, int32, int64, uint, uint8, uint16,
		uint32, uint64:
		return true, false
	}

	// allow types implementing TextMarshaler and TextUnmarshaler,
	// as they can be easily converted to string, which is supported by jsonapi
	return typ.Implements(reflect.TypeOf(new(encoding.TextMarshaler)).Elem()) &&
		typ.Implements(reflect.TypeOf(new(encoding.TextUnmarshaler)).Elem()), true
}

func (f *idField) jsonapiIdFields() []reflect.StructField {
	typ := f.fieldType
	if f.marshalling {
		typ = reflect.TypeOf("")
	}

	tag := fmt.Sprintf(`jsonapi:"primary,%s"`, f.schema.name)
	idField := reflect.StructField{
		Name: idFieldName,
		Type: typ,
		Tag:  reflect.StructTag(tag),
	}

	return []reflect.StructField{idField}
}

func (f *idField) pgIdFields() []reflect.StructField {
	tableNameField := reflect.StructField{
		Name: "TableName",
		Type: emptyStructType,
		// quote table name and alias as go-pg doesn't do it for aliases
		Tag: reflect.StructTag(fmt.Sprintf(`sql:"\"%s\",alias:\"%s\""`, f.schema.table, f.schema.alias)),
	}

	pgTag := fmt.Sprintf(`sql:"%s,pk`, IdFieldColumn)
	if isUUIDField(f.fieldType) {
		pgTag += `,type:uuid,default:uuid_generate_v4()`
	}
	pgTag += `"`

	idField := reflect.StructField{
		Name: idFieldName,
		Type: f.fieldType,
		Tag:  reflect.StructTag(pgTag),
	}

	return []reflect.StructField{tableNameField, idField}
}

func (f *idField) Writable() bool {
	return false
}

func (f *idField) Sortable() bool {
	return true
}

func (f *idField) Filterable() bool {
	return true
}

func (f *idField) JSONAPIName() string {
	return IdFieldJsonapiName
}

func (f *idField) ColumnName() string {
	return IdFieldColumn
}

func (f *idField) PGSelectColumn() string {
	return fmt.Sprintf("%s.%s", f.schema.alias, IdFieldColumn)
}

func (f *idField) PGFilterColumn() string {
	return f.PGSelectColumn()
}

func (f *idField) createInstance() schemaFieldInstance {
	return &idFieldInstance{
		field: f,
	}
}

func (f *idField) jsonapiFields() []reflect.StructField {
	return f.jsonapiF
}

func (f *idField) pgFields() []reflect.StructField {
	return f.pgF
}

func (f *idField) jsonapiJoinFields() []reflect.StructField {
	return f.jsonapiF
}

func (f *idField) pgJoinFields() []reflect.StructField {
	return f.pgF
}

// id fields are always valid
func (i *idFieldInstance) validate(*validator.Validate) error {
	return nil
}

// implements fieldInstance
type idFieldInstance struct {
	field *idField
	value interface{}
}

func (i *idFieldInstance) parentField() SchemaField {
	return i.field
}

func (i *idFieldInstance) sortValue() interface{} {
	return i.value
}

func (i *idFieldInstance) parseResourceModel(instance *resourceModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.parse(instance.value)
}

func (i *idFieldInstance) applyToResourceModel(instance *resourceModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.apply(instance.value)
}

func (i *idFieldInstance) parseJoinResourceModel(instance *resourceModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.parse(instance.value)
}

func (i *idFieldInstance) applyToJoinResourceModel(instance *resourceModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.apply(instance.value)
}

func (i *idFieldInstance) parseJsonapiModel(instance *jsonapiModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}

	v := instance.value
	if v.IsNil() {
		return
	}

	val := v.Elem().FieldByName(idFieldName).Interface()
	if i.field.marshalling {
		// unmarshal value from string
		val = reflect.New(i.field.fieldType).Elem().Interface()
		val.(encoding.TextUnmarshaler).UnmarshalText([]byte(val.(string)))
	}

	i.value = val
}

func (i *idFieldInstance) applyToJsonapiModel(instance *jsonapiModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}

	val := i.value
	if i.field.marshalling {
		// marshal the value into a string
		b, err := i.value.(encoding.TextMarshaler).MarshalText()
		if err != nil {
			panic(err)
		}
		val = string(b)
	}

	v := instance.value
	if v.IsNil() {
		panic(errNilPointer)
	}
	v.Elem().FieldByName(idFieldName).Set(reflect.ValueOf(val))
}

func (i *idFieldInstance) parseJoinJsonapiModel(instance *joinJsonapiModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.parse(instance.value)
}

func (i *idFieldInstance) applyToJoinJsonapiModel(instance *joinJsonapiModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.apply(instance.value)
}

func (i *idFieldInstance) parsePGModel(instance *pgModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.parse(instance.value)
}

func (i *idFieldInstance) applyToPGModel(instance *pgModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.apply(instance.value)
}

func (i *idFieldInstance) parseJoinPGModel(instance *joinPGModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.parse(instance.value)
}

func (i *idFieldInstance) applyToJoinPGModel(instance *joinPGModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.apply(instance.value)
}

// the id field is named "Id" in every representation,
// so the value of that field can be copied in any case.
func (i *idFieldInstance) parse(v *reflect.Value) {
	if !v.IsNil() {
		i.value = v.Elem().FieldByName(idFieldName).Interface()
	}
}

func (i *idFieldInstance) apply(v *reflect.Value) {
	if v.IsNil() {
		panic(errNilPointer)
	}
	v.Elem().FieldByName(idFieldName).Set(reflect.ValueOf(i.value))
}
