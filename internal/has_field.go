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

func newHasField(r Registry, schema *schema, f *reflect.StructField, fk string) (field, error) {
	base, err := newRelationField(r, schema, f)
	if err != nil {
		return nil, err
	}

	field := &hasField{
		relationField: base,
		fk:            fk,
	}

	// TODO: fail if there are invalid struct tag options

	return field, nil
}

// override this function to calculate topLevel pg fields on demand,
// i.e. after non-top-level pg fields were calculated for reference.
func (f *hasField) pgFields() ([]reflect.StructField, error) {
	if f.pgF != nil {
		return f.pgF, nil
	}

	f.pgF = pgHasFields(f)
	return f.pgF, nil
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
		f.relationField.createInstance(),
	}
}

type hasFieldInstance struct {
	*relationFieldInstance
}
