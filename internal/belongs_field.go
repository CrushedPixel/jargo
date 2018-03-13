package internal

import (
	"errors"
	"fmt"
	"github.com/c9s/inflect"
	"reflect"
)

var errInvalidBelongsToType = errors.New("invalid belongsTo field type. for a belongsToMany relation, use many2many")

type belongsToField struct {
	*relationField

	joinJsonapiFields []reflect.StructField
	joinPGFields      []reflect.StructField
}

func newBelongsToField(r SchemaRegistry, schema *Schema, f *reflect.StructField) SchemaField {
	base := newRelationField(r, schema, f)

	if base.collection {
		panic(errInvalidBelongsToType)
	}

	field := &belongsToField{
		relationField: base,
	}

	// TODO: fail if there are invalid struct tag options

	return field
}

func (f *belongsToField) ColumnName() string {
	return f.relationIdFieldColumn()
}

func (f *belongsToField) PGFilterColumn() string {
	// the column name for the relation id field generated by go-pg
	// is a snake_cased version of the id field name
	return fmt.Sprintf("%s.%s", f.schema.alias, f.relationIdFieldColumn())
}

// override this function to calculate topLevel pg fields on demand,
// i.e. after non-top-level pg fields were calculated for reference.
func (f *belongsToField) pgFields() []reflect.StructField {
	if f.pgF != nil {
		return f.pgF
	}

	f.pgF = pgBelongsToFields(f, false)
	return f.pgF
}

func (f *belongsToField) pgJoinFields() []reflect.StructField {
	if f.joinPGFields != nil {
		return f.joinPGFields
	}

	f.joinPGFields = pgBelongsToFields(f, true)
	return f.joinPGFields
}

// generates the pg fields for a belongsTo relation. Example:
// Owner *User `jargo:",has"
// =>
// OwnerId int64 // join model and full model
// Owner *User   // full model only
func pgBelongsToFields(f *belongsToField, joinField bool) []reflect.StructField {
	// every belongsTo association has a column containing
	// the id of the related resource
	tag := fmt.Sprintf(`sql:"%s`, f.relationIdFieldColumn())
	if !f.nullable {
		tag += ",notnull"
	}
	if f.sqlUnique {
		tag += ",unique"
	}
	tag += `"`

	idField := reflect.StructField{
		Name: f.relationIdFieldName(),
		Type: idFieldType,
		Tag:  reflect.StructTag(tag),
	}
	fields := []reflect.StructField{idField}

	if !joinField {
		// non-join fields contain a reference
		// to the full pg model of the relation,
		// so relation data can be fetched in queries
		field := reflect.StructField{
			Name: f.fieldName,
			Type: f.relationJoinPGFieldType(),
		}
		fields = append(fields, field)
	}

	return fields
}

func (f *belongsToField) relationIdFieldName() string {
	return fmt.Sprintf("%sId", f.fieldName)
}

func (f *belongsToField) relationIdFieldColumn() string {
	return inflect.Underscore(f.relationIdFieldName())
}

func (f *belongsToField) createInstance() schemaFieldInstance {
	return &belongsToFieldInstance{
		relationFieldInstance: f.relationField.createInstance(),
		field: f,
	}
}

type belongsToFieldInstance struct {
	*relationFieldInstance
	field *belongsToField
}

func (i *belongsToFieldInstance) parentField() SchemaField {
	return i.field
}

func (i *belongsToFieldInstance) sortValue() interface{} {
	val := i.values[0]
	// relations may be nil
	if val == nil {
		return nil
	}

	// return relation's id value
	for _, f := range val.fields {
		if idf, ok := f.(*idFieldInstance); ok {
			return idf.sortValue()
		}
	}

	panic("could not find relation's id field")
}

// parsePGModel parses the value of the pg relation struct field
// (e.g. *User) and stores it in i.values[0].
// If the relation struct field is nil, but the relation id field (e.g. UserId)
// is not zero, stores a new instance of the relation type with the id field set
// in i.values[0].
func (i *belongsToFieldInstance) parsePGModel(instance *pgModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.values = nil
	// do not parse nil models
	if instance.value.IsNil() {
		return
	}

	var schemaInstance *SchemaInstance
	pgModelInstance := instance.value.Elem().FieldByName(i.field.fieldName).Interface()
	schemaInstance = i.relationSchema.parseJoinPGModel(pgModelInstance)

	// if relation struct field is nil, but id field isn't,
	// create a new instance of the model and set its id field
	if schemaInstance == nil {
		id := instance.value.Elem().FieldByName(i.field.relationIdFieldName()).Int()
		if id != 0 {
			schemaInstance = i.relationSchema.createInstance()
			// set id field
			for _, f := range schemaInstance.fields {
				if idField, ok := f.(*idFieldInstance); ok {
					idField.value = id
				}
			}
		}
	}

	i.values = []*SchemaInstance{schemaInstance}
}

// sets the value of the pg relation id field (e.g. UserId) to the id value
// of the Schema instance stored in i.values[0]
func (i *belongsToFieldInstance) applyToPGModel(instance *pgModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	id, ok := i.relationId()
	if !ok {
		return
	}
	instance.value.Elem().FieldByName(i.field.relationIdFieldName()).SetInt(id)
	// apply relation instance to pg model field
	instance.value.Elem().FieldByName(i.field.fieldName).Set(reflect.ValueOf(i.values[0].toJoinPGModel()))
}

func (i *belongsToFieldInstance) parseJoinPGModel(instance *joinPGModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.values = nil
	// do not parse nil models
	if instance.value.IsNil() {
		return
	}

	// parse the id field for this relation
	idValue := instance.value.Elem().FieldByName(i.field.relationIdFieldName())

	// create an instance of the relation schema and set the id value
	relationInstance := i.relationSchema.createInstance()
	for _, f := range relationInstance.fields {
		if idField, ok := f.(*idFieldInstance); ok {
			idField.value = idValue.Int()
		}
	}

	// store the relation instance in values[0]
	i.values = []*SchemaInstance{relationInstance}
}

func (i *belongsToFieldInstance) applyToJoinPGModel(instance *joinPGModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}

	id, ok := i.relationId()
	if !ok {
		return
	}
	instance.value.Elem().FieldByName(i.field.relationIdFieldName()).SetInt(id)
}

func (i *belongsToFieldInstance) applyToJoinResourceModel(instance *resourceModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	if len(i.values) == 0 {
		return
	}

	v := i.values[0]
	// relations may be nil
	if v == nil {
		return
	}

	rmi := v.toJoinResourceModel()

	// if target field is not nullable,
	// dereference value pointer
	val := reflect.ValueOf(rmi)
	if !isNullable(i.field.fieldType) {
		val = val.Elem()
	}
	instance.value.Elem().FieldByName(i.field.fieldName).Set(val)
}

// relationId returns the id value of the relation.
func (i *belongsToFieldInstance) relationId() (int64, bool) {
	if len(i.values) == 0 {
		return 0, false
	}

	// extract id field from relation
	v := i.values[0]
	// relations may be nil
	if v == nil {
		return 0, false
	}

	var id *int64
	for _, f := range v.fields {
		if idField, ok := f.(*idFieldInstance); ok {
			id = &idField.value
		}
	}
	if id == nil {
		panic(errors.New("id field of related resource not found"))
	}
	return *id, true
}
