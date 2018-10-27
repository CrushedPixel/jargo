package internal

import (
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

	// the name of the field containing
	// table information in the pg struct tag
	pgTableNameFieldName = "TableName"
)

type FieldKind int

const (
	primitive FieldKind = iota
	uuid
	textMarshaler
)

// implements field
type idField struct {
	schema *Schema

	fieldType reflect.Type

	// the kind of the id field
	kind FieldKind

	jsonapiF []reflect.StructField
	pgF      []reflect.StructField
}

func newIdField(schema *Schema, f *reflect.StructField) SchemaField {
	idf := &idField{
		schema:    schema,
		fieldType: f.Type,
	}

	valid, kind := isValidIdField(idf.fieldType)
	if !valid {
		panic(errInvalidIdType)
	}
	idf.kind = kind

	// generate jsonapi and pg attribute fields
	idf.jsonapiF = idf.jsonapiIdFields()
	idf.pgF = idf.pgIdFields()

	// wrap id fields with uuid type in uuidIdField
	// for afterCreateTable hook
	if kind == uuid {
		return &uuidIdField{idf}
	}

	return idf
}

// isValidIdField returns whether typ is a valid type for an id field.
func isValidIdField(typ reflect.Type) (bool, FieldKind) {
	i := reflect.New(typ).Elem().Interface()

	switch i.(type) {
	case string, int, int8, int16, int32, int64, uint, uint8, uint16,
		uint32, uint64:
		return true, primitive
	}

	if (isTextMarshaler(typ) || pointerTypeIsTextMarshaler(typ)) &&
		(isTextUnmarshaler(typ) || pointerTypeIsTextUnmarshaler(typ)) {
		kind := textMarshaler
		if isUUIDField(typ) {
			kind = uuid
		}
		return true, kind
	}

	return false, 0
}

func (f *idField) jsonapiIdFields() []reflect.StructField {
	// jsonapi id field is always a string
	typ := reflect.TypeOf("")

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
		Name: pgTableNameFieldName,
		Type: emptyStructType,
		// quote table name and alias as go-pg doesn't do it for aliases
		Tag: reflect.StructTag(fmt.Sprintf(`sql:"\"%s\",alias:\"%s\""`, f.schema.table, f.schema.alias)),
	}

	pgTag := fmt.Sprintf(`sql:"%s,pk`, IdFieldColumn)
	if f.kind == uuid {
		pgTag += `,type:uuid,default:uuid_generate_v4()`
	}

	typ := f.fieldType
	if f.kind == textMarshaler {
		// text marshaler types have to be stored as strings
		typ = reflect.TypeOf("")
	}

	pgTag += `"`

	idField := reflect.StructField{
		Name: idFieldName,
		Type: typ,
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

func (f *idField) typ() reflect.Type {
	return f.fieldType
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
	i.parseJsonapi(instance.value)
}

func (i *idFieldInstance) applyToJsonapiModel(instance *jsonapiModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.applyJsonapi(instance.value)
}

func (i *idFieldInstance) parseJoinJsonapiModel(instance *joinJsonapiModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.parseJsonapi(instance.value)
}

func (i *idFieldInstance) applyToJoinJsonapiModel(instance *joinJsonapiModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.applyJsonapi(instance.value)
}

func (i *idFieldInstance) parsePGModel(instance *pgModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.parsePG(instance.value)
}

func (i *idFieldInstance) applyToPGModel(instance *pgModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.applyPG(instance.value)
}

func (i *idFieldInstance) parseJoinPGModel(instance *joinPGModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.parsePG(instance.value)
}

func (i *idFieldInstance) applyToJoinPGModel(instance *joinPGModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.applyPG(instance.value)
}

// the id field is named "Id" in every representation,
// so the value of that field can be copied in any case.
func (i *idFieldInstance) parse(v *reflect.Value) {
	if v.IsNil() {
		panic(errNilPointer)
	}
	i.value = v.Elem().FieldByName(idFieldName).Interface()
}

func (i *idFieldInstance) apply(v *reflect.Value) {
	if v.IsNil() {
		panic(errNilPointer)
	}
	v.Elem().FieldByName(idFieldName).Set(reflect.ValueOf(i.value))
}

func (i *idFieldInstance) parseJsonapi(v *reflect.Value) {
	if v.IsNil() {
		panic(errNilPointer)
	}
	i.value = StringToId(v.Elem().FieldByName(idFieldName).String(), i.field.fieldType)
}

func (i *idFieldInstance) applyJsonapi(v *reflect.Value) {
	if v.IsNil() {
		panic(errNilPointer)
	}
	v.Elem().FieldByName(idFieldName).SetString(IdToString(i.value))
}

func (i *idFieldInstance) parsePG(v *reflect.Value) {
	if v.IsNil() {
		panic(errNilPointer)
	}

	if i.field.kind == textMarshaler {
		// unmarshal the PG field's string value into the appropiate type
		i.value = StringToTextUnmarshaler(v.Elem().FieldByName(idFieldName).String(), i.field.fieldType)
	} else {
		i.value = v.Elem().FieldByName(idFieldName).Interface()
	}
}

func (i *idFieldInstance) applyPG(v *reflect.Value) {
	if v.IsNil() {
		panic(errNilPointer)
	}

	if i.field.kind == textMarshaler {
		// marshal the id field's value into a string
		v.Elem().FieldByName(idFieldName).SetString(IdToString(i.value))
	} else {
		v.Elem().FieldByName(idFieldName).Set(reflect.ValueOf(i.value))
	}
}
