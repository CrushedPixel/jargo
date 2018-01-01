package internal

import (
	"reflect"
	"fmt"
	"errors"
)

var errInvalidBelongsToType = errors.New("invalid belongsTo field type. for a belongsToMany relation, use many2many")

type belongsToField struct {
	*relationField

	joinJsonapiFields []reflect.StructField
	joinPGFields      []reflect.StructField
}

func newBelongsToField(r Registry, schema *schema, f *reflect.StructField, fk string) (field, error) {
	base, err := newRelationField(r, schema, f)
	if err != nil {
		return nil, err
	}

	if base.collection {
		return nil, errInvalidBelongsToType
	}

	field := &belongsToField{
		relationField: base,
	}

	// TODO: fail if there are invalid struct tag options

	return field, nil
}

// override this function to calculate topLevel pg fields on demand,
// i.e. after non-top-level pg fields were calculated for reference.
func (f *belongsToField) pgFields() ([]reflect.StructField, error) {
	if f.pgF != nil {
		return f.pgF, nil
	}

	f.pgF = pgBelongsToFields(f, false)
	return f.pgF, nil
}

func (f *belongsToField) pgJoinFields() ([]reflect.StructField, error) {
	if f.joinPGFields != nil {
		return f.joinPGFields, nil
	}

	f.joinPGFields = pgBelongsToFields(f, true)
	return f.joinPGFields, nil
}

// generates the pg fields for a belongsTo relation. Example:
// Owner *User `jargo:",hasMany"
// =>
// OwnerId int64 // join model and full model
// Owner *User   // full model only
func pgBelongsToFields(f *belongsToField, joinField bool) []reflect.StructField {
	// every belongsTo association has a column containing
	// the id of the related resource
	idField := reflect.StructField{
		Name: f.relationIdFieldName(),
		Type: idFieldType,
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

func (f *belongsToField) createInstance() fieldInstance {
	return &belongsToFieldInstance{
		relationFieldInstance: f.relationField.createInstance(),
		field:                 f,
	}
}

type belongsToFieldInstance struct {
	*relationFieldInstance
	field *belongsToField
}

func (i *belongsToFieldInstance) parsePGModel(instance *pgModelInstance) error {
	println("ParsePGModel")
	return nil // TODO
}

func (i *belongsToFieldInstance) applyToPGModel(instance *pgModelInstance) error {
	println("ApplyToPGModel")
	return nil // TODO
}

func (i *belongsToFieldInstance) parseJoinPGModel(instance *joinPGModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}

	// do not parse nil models
	if instance.value.IsNil() {
		return nil
	}

	// get id field value
	id := instance.value.Elem().FieldByName(i.field.relationIdFieldName()).Int()
	println("id field value", id) // TODO

	return nil
}

func (i *belongsToFieldInstance) applyToJoinPGModel(instance *joinPGModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}

	/* // TODO
	v := i.values[0]
	if v != nil {
		println("HERE")
		// get id value

		// apply model id to id field
		for _, f := range v.(*schemaInstance).fields {
			if idf, ok := f.(*idFieldInstance); ok {
				idf.value =
			}
		}
		panic(errors.New("id field instance not found"))

	}
	*/

	return nil
}