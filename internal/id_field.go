package internal

import (
	"reflect"
	"fmt"
	"errors"
)

var (
	errMismatchingSchema = errors.New("mismatching schema")
	emptyStructType      = reflect.TypeOf(new(struct{}))
)

const (
	idFieldColumn      = "id"
	idFieldJsonapiName = "id"
)

// implements field
type idField struct {
	schema *schema

	jsonapiF []reflect.StructField
	pgF      []reflect.StructField
}

func newIdField(schema *schema) (field, error) {
	f := &idField{
		schema:   schema,
		jsonapiF: jsonapiIdFields(schema),
		pgF:      pgIdFields(schema),
	}
	return f, nil
}

func jsonapiIdFields(schema *schema) []reflect.StructField {
	tag := fmt.Sprintf(`jsonapi:"primary,%s"`, schema.name)
	idField := reflect.StructField{
		Name: idFieldName,
		Type: idFieldType,
		Tag:  reflect.StructTag(tag),
	}

	return []reflect.StructField{idField}
}

func pgIdFields(schema *schema) []reflect.StructField {
	tableNameField := reflect.StructField{
		Name: "TableName",
		Type: emptyStructType,
		// quote table name and alias as go-pg doesn't do it for aliases
		Tag: reflect.StructTag(fmt.Sprintf(`sql:"\"%s\",alias:\"%s\""`, schema.table, schema.alias)),
	}

	idField := reflect.StructField{
		Name: idFieldName,
		Type: idFieldType,
		Tag:  reflect.StructTag(fmt.Sprintf(`sql:"%s,pk"`, idFieldColumn)),
	}

	return []reflect.StructField{tableNameField, idField}
}

func (f *idField) writable() bool {
	return false
}

func (f *idField) sortable() bool {
	return true
}

func (f *idField) filterable() bool {
	return true
}

func (f *idField) jsonapiName() string {
	return idFieldJsonapiName
}

func (f *idField) pgSelectColumn() string {
	return fmt.Sprintf("%s.%s", f.schema.alias, idFieldColumn)
}

func (f *idField) pgFilterColumn() string {
	return f.pgSelectColumn()
}

func (f *idField) createInstance() fieldInstance {
	return &idFieldInstance{
		field: f,
	}
}

func (f *idField) jsonapiFields() ([]reflect.StructField, error) {
	return f.jsonapiF, nil
}

func (f *idField) pgFields() ([]reflect.StructField, error) {
	return f.pgF, nil
}

func (f *idField) jsonapiJoinFields() ([]reflect.StructField, error) {
	return f.jsonapiF, nil
}

func (f *idField) pgJoinFields() ([]reflect.StructField, error) {
	return f.pgF, nil
}

// id fields are always valid
func (i *idFieldInstance) validate() error {
	return nil
}

// implements fieldInstance
type idFieldInstance struct {
	field *idField
	value int64
}

func (i *idFieldInstance) parentField() field {
	return i.field
}

func (i *idFieldInstance) parseResourceModel(instance *resourceModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	return i.parse(instance.value)
}

func (i *idFieldInstance) applyToResourceModel(instance *resourceModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	return i.apply(instance.value)
}

func (i *idFieldInstance) parseJoinResourceModel(instance *resourceModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	return i.parse(instance.value)
}

func (i *idFieldInstance) applyToJoinResourceModel(instance *resourceModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	return i.apply(instance.value)
}

func (i *idFieldInstance) parseJsonapiModel(instance *jsonapiModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	return i.parse(instance.value)
}

func (i *idFieldInstance) applyToJsonapiModel(instance *jsonapiModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	return i.apply(instance.value)
}

func (i *idFieldInstance) parseJoinJsonapiModel(instance *joinJsonapiModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	return i.parse(instance.value)
}

func (i *idFieldInstance) applyToJoinJsonapiModel(instance *joinJsonapiModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	return i.apply(instance.value)
}

func (i *idFieldInstance) parsePGModel(instance *pgModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	return i.parse(instance.value)
}

func (i *idFieldInstance) applyToPGModel(instance *pgModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	return i.apply(instance.value)
}

func (i *idFieldInstance) parseJoinPGModel(instance *joinPGModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	return i.parse(instance.value)
}

func (i *idFieldInstance) applyToJoinPGModel(instance *joinPGModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	return i.apply(instance.value)
}

// the id field is named "Id" in every representation,
// so the value of that field can be copied in any case.
func (i *idFieldInstance) parse(v *reflect.Value) error {
	if v.IsNil() {
		return nil
	}
	i.value = v.Elem().FieldByName(idFieldName).Int()
	return nil
}

func (i *idFieldInstance) apply(v *reflect.Value) error {
	if v.IsNil() {
		panic(errors.New("struct pointer must not be nil"))
	}
	v.Elem().FieldByName(idFieldName).SetInt(i.value)
	return nil
}
