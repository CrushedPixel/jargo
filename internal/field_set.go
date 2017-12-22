package internal

import (
	"errors"
	"fmt"
	"reflect"
	"github.com/go-pg/pg/orm"
	"crushedpixel.net/jargo/api"
)

type FieldSet struct {
	resource *Resource
	fields   []*resourceField
}

func (fs *FieldSet) ApplyToQuery(q *orm.Query) {
	// always select the Id field
	column(q, fs.resource, primaryFieldColumn)

	// select all columns required by the fieldSet
	for _, f := range fs.fields {
		switch f.definition.typ {
		case attribute:
			column(q, fs.resource, f.definition.column)
		case belongsTo, has, many2many:
			q.Column(f.definition.structField.Name)
		}
	}
}

func column(q *orm.Query, resource *Resource, column string) {
	q.Column(fmt.Sprintf("%s.%s", resource.definition.alias, column))
}

func (fs *FieldSet) Resource() api.Resource {
	return fs.resource
}

// copy all of the FieldSet's values from source to the target struct.
// takes struct pointer values.
func (fs *FieldSet) applyValues(source reflect.Value, target reflect.Value) {
	// always copy the id field
	value := source.Elem().FieldByName(primaryFieldName)
	target.Elem().FieldByName(primaryFieldName).Set(value)
	for _, field := range fs.fields {
		fieldName := field.definition.structField.Name
		sourceField := source.Elem().FieldByName(fieldName)
		targetField := target.Elem().FieldByName(fieldName)
		if !targetField.IsValid() {
			continue
		}

		targetField.Set(sourceField)
	}
}

func (fs *FieldSet) pgFields() []reflect.StructField {
	fields := make([]reflect.StructField, 0)

	for _, f := range fs.fields {
		fields = append(fields, f.pgFields...)
	}

	return fields
}

func (fs *FieldSet) jsonapiFields() []reflect.StructField {
	fields := make([]reflect.StructField, 0)

	for _, f := range fs.fields {
		fields = append(fields, f.jsonapiFields...)
	}

	return fields
}

func allFields(resource *Resource) *FieldSet {
	return &FieldSet{
		resource: resource,
		fields:   resource.fields,
	}
}

// returns a FieldSet containing all fields of a resource that are not marked readonly
func settableFields(resource *Resource) *FieldSet {
	fields := make([]*resourceField, 0)

	for _, f := range resource.fields {
		if !f.definition.readonly {
			fields = append(fields, f)
		}
	}

	return &FieldSet{
		resource: resource,
		fields:   fields,
	}
}

// parses a comma-separated list of field names into a FieldSet for a given Resource
func newFieldSet(resource *Resource, names []string) (*FieldSet, error) {
	fields := make([]*resourceField, 0)

	for _, name := range names {
		var field *resourceField
		// find the resourceField with matching jsonapi name
		for _, f := range resource.fields {
			if name == f.definition.name {
				field = f
			}
		}

		if field == nil {
			return nil, errors.New(fmt.Sprintf(`unknown field parameter: "%s"`, name))
		}

		fields = append(fields, field)
	}

	fieldSet := &FieldSet{
		resource: resource,
		fields:   fields,
	}
	return fieldSet, nil
}
