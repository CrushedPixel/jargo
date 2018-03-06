package internal

import (
	"errors"
	"fmt"
	"gopkg.in/go-playground/validator.v9"
	"reflect"
)

var (
	errNilPointer        = errors.New("resource must not be nil")
	errMismatchingSchema = errors.New("mismatching schema")
	emptyStructType      = reflect.TypeOf(new(struct{}))
)

const (
	IdFieldColumn      = "id"
	IdFieldJsonapiName = "id"
)

// implements field
type idField struct {
	schema *Schema

	jsonapiF []reflect.StructField
	pgF      []reflect.StructField
}

func newIdField(schema *Schema) SchemaField {
	f := &idField{
		schema:   schema,
		jsonapiF: jsonapiIdFields(schema),
		pgF:      pgIdFields(schema),
	}
	return f
}

func jsonapiIdFields(schema *Schema) []reflect.StructField {
	tag := fmt.Sprintf(`jsonapi:"primary,%s"`, schema.name)
	idField := reflect.StructField{
		Name: idFieldName,
		Type: idFieldType,
		Tag:  reflect.StructTag(tag),
	}

	return []reflect.StructField{idField}
}

func pgIdFields(schema *Schema) []reflect.StructField {
	tableNameField := reflect.StructField{
		Name: "TableName",
		Type: emptyStructType,
		// quote table name and alias as go-pg doesn't do it for aliases
		Tag: reflect.StructTag(fmt.Sprintf(`sql:"\"%s\",alias:\"%s\""`, schema.table, schema.alias)),
	}

	idField := reflect.StructField{
		Name: idFieldName,
		Type: idFieldType,
		Tag:  reflect.StructTag(fmt.Sprintf(`sql:"%s,pk"`, IdFieldColumn)),
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
	value int64
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
	i.parse(instance.value)
}

func (i *idFieldInstance) applyToJsonapiModel(instance *jsonapiModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.apply(instance.value)
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
		i.value = v.Elem().FieldByName(idFieldName).Int()
	}
}

func (i *idFieldInstance) apply(v *reflect.Value) {
	if v.IsNil() {
		panic(errNilPointer)
	}
	v.Elem().FieldByName(idFieldName).SetInt(i.value)
}
