package internal

import (
	"reflect"
	"fmt"
)

type hasField struct {
	*relationField

	joinPGFields []reflect.StructField
	fk           string
}

func newHasField(r Registry, schema *schema, f *reflect.StructField, fk string) field {
	base := newRelationField(r, schema, f)

	if fk == "" {
		// the internal pg model struct types are unnamed,
		// because they are generated at runtime.
		// therefore, we need to provide go-pg with a foreign key
		// as it can't fall back to the type name.
		// we use the original resource model type name
		// as default foreign key on has relations.
		fk = schema.resourceModelType.Name()
	}
	field := &hasField{
		relationField: base,
		fk:            fk,
	}

	// TODO: fail if there are invalid struct tag options

	return field
}

// override this function to calculate topLevel pg fields on demand,
// i.e. after non-top-level pg fields were calculated for reference.
func (f *hasField) pgFields() []reflect.StructField {
	if f.pgF != nil {
		return f.pgF
	}

	f.pgF = pgHasFields(f)
	return f.pgF
}

func pgHasFields(f *hasField) []reflect.StructField {
	field := reflect.StructField{
		Name: f.fieldName,
		Type: f.relationJoinPGFieldType(),
	}

	if f.fk != "" {
		field.Tag = reflect.StructTag(fmt.Sprintf(`pg:",fk:%s"`, f.fk))
	}

	return []reflect.StructField{field}
}

func (f *hasField) createInstance() fieldInstance {
	return &hasFieldInstance{
		relationFieldInstance: f.relationField.createInstance(),
		field:                 f,
	}
}

type hasFieldInstance struct {
	*relationFieldInstance
	field *hasField
}

func (i *hasFieldInstance) parentField() field {
	return i.field
}
