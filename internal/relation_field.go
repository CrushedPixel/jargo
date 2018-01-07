package internal

import (
	"errors"
	"fmt"
	"github.com/crushedpixel/jargo/api"
	"gopkg.in/go-playground/validator.v9"
	"reflect"
)

func errInvalidRelationFieldType(p reflect.Type) error {
	return errors.New(fmt.Sprintf("invalid type for relation field: %s", p))
}

type relationField struct {
	*baseField

	registry Registry

	relationType reflect.Type // struct type of relation
	collection   bool         // whether it's a to-many-relation
}

func newRelationField(r Registry, schema *schema, f *reflect.StructField) *relationField {
	base := newBaseField(schema, f)

	// validate field type
	typ, collection := getRelationType(f.Type)
	if typ == nil {
		panic(errInvalidRelationFieldType(f.Type))
	}

	field := &relationField{
		baseField:    base,
		registry:     r,
		relationType: typ,
		collection:   collection,
	}

	return field
}

func getRelationType(typ reflect.Type) (reflect.Type, bool) {
	collection := false
	if typ.Kind() == reflect.Slice {
		typ = typ.Elem()
		collection = true
	}

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	} else {
		return nil, false
	}

	if typ.Kind() != reflect.Struct {
		return nil, false
	}

	return typ, collection
}

func (f *relationField) pgSelectColumn() string {
	return f.fieldName
}

func (f *relationField) pgFilterColumn() string {
	panic("unsupported operation")
}

// override this function to calculate topLevel jsonapi fields on demand,
// i.e. after non-top-level jsonapi fields were calculated for reference.
func (f *relationField) jsonapiFields() []reflect.StructField {
	if f.jsonapiF != nil {
		return f.jsonapiF
	}

	f.jsonapiF = jsonapiRelationFields(f)
	return f.jsonapiF
}

func (f *relationField) jsonapiJoinFields() []reflect.StructField {
	// relations are not present in join fields to avoid infinite recursion
	return []reflect.StructField{}
}

func (f *relationField) pgJoinFields() []reflect.StructField {
	// relations are not present in join fields to avoid infinite recursion
	return []reflect.StructField{}
}

func jsonapiRelationFields(f *relationField) []reflect.StructField {
	if f.name == unexportedFieldName {
		return []reflect.StructField{}
	}

	tag := fmt.Sprintf(`jsonapi:"relation,%s`, f.name)
	if f.jsonapiOmitempty {
		tag += `,omitempty`
	}
	tag += `"`

	// register relation schema
	f.registry.registerResource(f.relationType)

	typ := reflect.PtrTo(f.registry[f.relationType].joinJsonapiModelType)
	if f.collection {
		typ = reflect.SliceOf(typ)
	}

	field := reflect.StructField{
		Name: f.fieldName,
		Type: typ,
		Tag:  reflect.StructTag(tag),
	}
	return []reflect.StructField{field}
}

func (f *relationField) relationJoinJsonapiFieldType() reflect.Type {
	t := reflect.PtrTo(f.registry[f.relationType].joinJsonapiModelType)
	if f.collection {
		t = reflect.SliceOf(t)
	}
	return t
}

func (f *relationField) relationJoinPGFieldType() reflect.Type {
	t := reflect.PtrTo(f.registry[f.relationType].joinPGModelType)
	if f.collection {
		t = reflect.SliceOf(t)
	}
	return t
}

func (f *relationField) createInstance() *relationFieldInstance {
	relation, err := f.registry.RegisterResource(f.relationType)
	if err != nil {
		panic(err)
	}

	return &relationFieldInstance{
		field:          f,
		relationSchema: relation,
	}
}

type relationFieldInstance struct {
	field          *relationField
	relationSchema api.Schema // schema of the related resource
	values         []api.SchemaInstance
}

func (i *relationFieldInstance) parseResourceModel(instance *resourceModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	// do not parse nil models
	if instance.value.IsNil() {
		return
	}

	i.values = make([]api.SchemaInstance, 0)
	val := instance.value.Elem().FieldByName(i.field.fieldName)
	if i.field.collection {
		for x := 0; x < val.Len(); x++ {
			v := val.Index(x) // struct pointer value of related resource model
			i.values = append(i.values, i.relationSchema.ParseJoinResourceModel(v.Interface()))
		}
	} else {
		i.values = append(i.values, i.relationSchema.ParseJoinResourceModel(val.Interface()))
	}
}

func (i *relationFieldInstance) applyToResourceModel(instance *resourceModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}

	val := instance.value.Elem().FieldByName(i.field.fieldName)
	if i.field.collection {
		l := len(i.values)
		values := reflect.MakeSlice(i.field.fieldType, l, l)
		for x := 0; x < l; x++ {
			v := i.values[x]
			if v != nil {
				values.Index(x).Set(reflect.ValueOf(v.ToJoinResourceModel()))
			}
		}
		val.Set(values)
	} else {
		v := i.values[0]
		if v != nil {
			val.Set(reflect.ValueOf(v.ToJoinResourceModel()))
		}
	}
}

func (i *relationFieldInstance) parseJoinResourceModel(instance *resourceModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	// relations do not have their own relations set,
	// to avoid infinite recursion
}

func (i *relationFieldInstance) applyToJoinResourceModel(instance *resourceModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	// relations do not have their own relations set,
	// to avoid infinite recursion
}

func (i *relationFieldInstance) parseJsonapiModel(instance *jsonapiModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}

	// do not parse nil models
	if instance.value.IsNil() {
		return
	}

	i.values = make([]api.SchemaInstance, 0)
	val := instance.value.Elem().FieldByName(i.field.fieldName)
	if i.field.collection {
		for x := 0; x < val.Len(); x++ {
			v := val.Index(x) // struct pointer value of related resource model
			i.values = append(i.values, i.relationSchema.ParseJoinJsonapiModel(v.Interface()))
		}
	} else {
		i.values = append(i.values, i.relationSchema.ParseJoinJsonapiModel(val.Interface()))
	}
}

func (i *relationFieldInstance) applyToJsonapiModel(instance *jsonapiModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}

	val := instance.value.Elem().FieldByName(i.field.fieldName)
	if i.field.collection {
		l := len(i.values)
		values := reflect.MakeSlice(i.field.relationJoinJsonapiFieldType(), l, l)
		for x := 0; x < l; x++ {
			v := i.values[x]
			if v != nil {
				values.Index(x).Set(reflect.ValueOf(v.ToJoinJsonapiModel()))
			}
		}
		val.Set(values)
	} else {
		v := i.values[0]
		if v != nil {
			val.Set(reflect.ValueOf(v.ToJoinJsonapiModel()))
		}
	}
}

func (i *relationFieldInstance) parseJoinJsonapiModel(instance *joinJsonapiModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	// only the id field exists in join jsonapi models
}

func (i *relationFieldInstance) applyToJoinJsonapiModel(instance *joinJsonapiModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	// only the id field exists in join jsonapi models
}

func (i *relationFieldInstance) parsePGModel(instance *pgModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}

	// do not parse nil models
	if instance.value.IsNil() {
		return
	}

	i.values = make([]api.SchemaInstance, 0)
	val := instance.value.Elem().FieldByName(i.field.fieldName)
	if i.field.collection {
		for x := 0; x < val.Len(); x++ {
			v := val.Index(x) // struct pointer value of related resource model
			i.values = append(i.values, i.relationSchema.ParseJoinPGModel(v.Interface()))
		}
	} else {
		i.values = append(i.values, i.relationSchema.ParseJoinPGModel(val.Interface()))
	}
}

func (i *relationFieldInstance) applyToPGModel(instance *pgModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}

	val := instance.value.Elem().FieldByName(i.field.fieldName)
	if i.field.collection {
		l := len(i.values)
		values := reflect.MakeSlice(i.field.relationJoinPGFieldType(), l, l)
		for x := 0; x < l; x++ {
			v := i.values[x]
			if v != nil {
				values.Index(x).Set(reflect.ValueOf(v.ToJoinPGModel()))
			}
		}
		val.Set(values)
	} else {
		v := i.values[0]
		if v != nil {
			val.Set(reflect.ValueOf(v.ToJoinPGModel()))
		}
	}
}

func (i *relationFieldInstance) parseJoinPGModel(instance *joinPGModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	// join pg models do not have relation fields
}

func (i *relationFieldInstance) applyToJoinPGModel(instance *joinPGModelInstance) {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	// join pg models do not have relation fields
}

func (i *relationFieldInstance) validate(*validator.Validate) error {
	// relations are not validated
	return nil
}
