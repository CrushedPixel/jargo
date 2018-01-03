package internal

import (
	"reflect"
	"fmt"
	"errors"
	"crushedpixel.net/jargo/api"
	"github.com/c9s/inflect"
)

var errInvalidBelongsToType = errors.New("invalid belongsTo field type. for a belongsToMany relation, use many2many")

type belongsToField struct {
	*relationField

	joinJsonapiFields []reflect.StructField
	joinPGFields      []reflect.StructField
}

func newBelongsToField(r Registry, schema *schema, f *reflect.StructField) field {
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

func (f *belongsToField) pgFilterColumn() string {
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
// Owner *User `jargo:",hasMany"
// =>
// OwnerId int64 // join model and full model
// Owner *User   // full model only
func pgBelongsToFields(f *belongsToField, joinField bool) []reflect.StructField {
	// every belongsTo association has a column containing
	// the id of the related resource
	tag := fmt.Sprintf(`sql:"%s`, f.relationIdFieldColumn())
	if f.sqlNotnull {
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

func (i *belongsToFieldInstance) parentField() field {
	return i.field
}

// parses the value of the pg relation struct field (e.g. *User) and stores it
// in i.values[0]
func (i *belongsToFieldInstance) parsePGModel(instance *pgModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	i.values = nil
	// do not parse nil models
	if instance.value.IsNil() {
		return
	}

	pgModelInstance := instance.value.Elem().FieldByName(i.field.fieldName).Interface()
	i.values = []api.SchemaInstance{i.relationSchema.ParseJoinPGModel(pgModelInstance)}
}

// sets the value of the pg relation id field (e.g. UserId) to the id value
// of the schema instance stored in i.values[0]
func (i *belongsToFieldInstance) applyToPGModel(instance *pgModelInstance) {
	if len(i.values) == 0 {
		return
	}

	// extract id field from relation and apply value
	// to pg id field
	v := i.values[0].(*schemaInstance)
	var id *int64
	for _, f := range v.fields {
		if idField, ok := f.(*idFieldInstance); ok {
			id = &idField.value
		}
	}
	if id == nil {
		panic(errors.New("id field of related resource not found"))
	}
	instance.value.Elem().FieldByName(i.field.relationIdFieldName()).SetInt(*id)

	// apply relation instance to pg model field
	instance.value.Elem().FieldByName(i.field.fieldName).Set(reflect.ValueOf(v.ToJoinPGModel()))
}
