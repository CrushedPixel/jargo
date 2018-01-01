package internal

import (
	"reflect"
	"fmt"
	"errors"
	"crushedpixel.net/jargo/api"
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

func newRelationField(r Registry, schema *schema, f *reflect.StructField) (*relationField, error) {
	base, err := newBaseField(schema, f)
	if err != nil {
		return nil, err
	}

	// validate field type
	typ, collection := getRelationType(f.Type)
	if typ == nil {
		return nil, errInvalidRelationFieldType(f.Type)
	}

	field := &relationField{
		baseField:    base,
		registry:     r,
		relationType: typ,
		collection:   collection,
	}

	return field, nil
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
func (f *relationField) jsonapiFields() ([]reflect.StructField, error) {
	if f.jsonapiF != nil {
		return f.jsonapiF, nil
	}

	var err error
	f.jsonapiF, err = jsonapiRelationFields(f)
	if err != nil {
		return nil, err
	}
	return f.jsonapiF, nil
}

func (f *relationField) jsonapiJoinFields() ([]reflect.StructField, error) {
	// relations are not present in join fields to avoid infinite recursion
	return []reflect.StructField{}, nil
}

func (f *relationField) pgJoinFields() ([]reflect.StructField, error) {
	// relations are not present in join fields to avoid infinite recursion
	return []reflect.StructField{}, nil
}

func jsonapiRelationFields(f *relationField) ([]reflect.StructField, error) {
	if f.name == unexportedFieldName {
		return []reflect.StructField{}, nil
	}

	tag := fmt.Sprintf(`jsonapi:"relation,%s`, f.name)
	if f.jsonapiOmitempty {
		tag += `,omitempty`
	}
	tag += `"`

	// register relation schema
	err := f.registry.registerResource(f.relationType)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error registering related resource: %s", err))
	}

	typ := reflect.PtrTo(f.registry[f.relationType].joinJsonapiModelType)
	if f.collection {
		typ = reflect.SliceOf(typ)
	}

	field := reflect.StructField{
		Name: f.fieldName,
		Type: typ,
		Tag:  reflect.StructTag(tag),
	}

	return []reflect.StructField{field}, nil
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
	// relation has already been registered when creating the
	// schema field, so no error should be thrown here
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

func (i *relationFieldInstance) parseResourceModel(instance *resourceModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}

	// do not parse nil models
	if instance.value.IsNil() {
		return nil
	}

	i.values = make([]api.SchemaInstance, 0)

	val := instance.value.Elem().FieldByName(i.field.fieldName)
	if i.field.collection {
		for x := 0; x < val.Len(); x++ {
			v := val.Index(x) // struct pointer value of related resource model
			r, err := i.relationSchema.ParseJoinResourceModel(v.Interface())
			if err != nil {
				return err
			}
			i.values = append(i.values, r)
		}
	} else {
		v := val
		r, err := i.relationSchema.ParseJoinResourceModel(v.Interface())
		if err != nil {
			return err
		}
		i.values = append(i.values, r)
	}

	return nil
}

func (i *relationFieldInstance) applyToResourceModel(instance *resourceModelInstance) error {
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
				rm, err := v.ToJoinResourceModel()
				if err != nil {
					return err
				}
				values.Index(x).Set(reflect.ValueOf(rm))
			}
		}
		val.Set(values)
	} else {
		v := i.values[0]
		if v != nil {
			rm, err := v.ToJoinResourceModel()
			if err != nil {
				return err
			}
			val.Set(reflect.ValueOf(rm))
		}
	}

	return nil
}

func (i *relationFieldInstance) parseJoinResourceModel(instance *resourceModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	// relations do not have their own relations set,
	// to avoid infinite recursion
	return nil
}

func (i *relationFieldInstance) applyToJoinResourceModel(instance *resourceModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	// relations do not have their own relations set,
	// to avoid infinite recursion
	return nil
}

func (i *relationFieldInstance) parseJsonapiModel(instance *jsonapiModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}

	// do not parse nil models
	if instance.value.IsNil() {
		return nil
	}

	i.values = make([]api.SchemaInstance, 0)

	val := instance.value.Elem().FieldByName(i.field.fieldName)
	if i.field.collection {
		for x := 0; x < val.Len(); x++ {
			v := val.Index(x) // struct pointer value of related resource model
			r, err := i.relationSchema.ParseJoinJsonapiModel(v.Interface())
			if err != nil {
				return err
			}
			i.values = append(i.values, r)
		}
	} else {
		v := val
		r, err := i.relationSchema.ParseJoinJsonapiModel(v.Interface())
		if err != nil {
			return err
		}
		i.values = append(i.values, r)
	}

	return nil
}

func (i *relationFieldInstance) applyToJsonapiModel(instance *jsonapiModelInstance) error {
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
				rm, err := v.ToJoinJsonapiModel()
				if err != nil {
					return err
				}
				values.Index(x).Set(reflect.ValueOf(rm))
			}
		}
		val.Set(values)
	} else {
		v := i.values[0]
		if v != nil {
			rm, err := v.ToJoinJsonapiModel()
			if err != nil {
				return err
			}
			val.Set(reflect.ValueOf(rm))
		}
	}

	return nil
}

func (i *relationFieldInstance) parseJoinJsonapiModel(instance *joinJsonapiModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	// only the id field exists in join jsonapi models
	return nil
}

func (i *relationFieldInstance) applyToJoinJsonapiModel(instance *joinJsonapiModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	// only the id field exists in join jsonapi models
	return nil
}

func (i *relationFieldInstance) parsePGModel(instance *pgModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}

	// do not parse nil models
	if instance.value.IsNil() {
		return nil
	}

	i.values = make([]api.SchemaInstance, 0)

	val := instance.value.Elem().FieldByName(i.field.fieldName)
	if i.field.collection {
		for x := 0; x < val.Len(); x++ {
			v := val.Index(x) // struct pointer value of related resource model
			r, err := i.relationSchema.ParseJoinPGModel(v.Interface())
			if err != nil {
				return err
			}
			i.values = append(i.values, r)
		}
	} else {
		v := val
		r, err := i.relationSchema.ParseJoinPGModel(v.Interface())
		if err != nil {
			return err
		}
		i.values = append(i.values, r)
	}

	return nil
}

func (i *relationFieldInstance) applyToPGModel(instance *pgModelInstance) error {
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
				rm, err := v.ToJoinPGModel()
				if err != nil {
					return err
				}
				values.Index(x).Set(reflect.ValueOf(rm))
			}
		}
		val.Set(values)
	} else {
		v := i.values[0]
		if v != nil {
			rm, err := v.ToJoinPGModel()
			if err != nil {
				return err
			}
			val.Set(reflect.ValueOf(rm))
		}
	}

	return nil
}

func (i *relationFieldInstance) parseJoinPGModel(instance *joinPGModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	// join pg models do not have relation fields.
	return nil
}

func (i *relationFieldInstance) applyToJoinPGModel(instance *joinPGModelInstance) error {
	if i.field.schema != instance.schema {
		panic(errMismatchingSchema)
	}
	// join pg models do not have relation fields.
	return nil
}
