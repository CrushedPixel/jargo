package jargo

import (
	"reflect"
	"github.com/go-pg/pg/orm"
)

type ResourceFields map[string]*ResourceField

type ResourceField struct {
	StructField *reflect.StructField
	Name        string            // jsonapi attribute name
	Type        ResourceFieldType // jsonapi field type
	PGField     *orm.Field        // go-pg field
	Settings    *FieldSettings    // jargo field settings
}

type FieldSettings struct {
	AllowFiltering bool // if true, filtering by this field is allowed
	AllowSorting   bool // if true, sorting by this field is allowed
	Readonly       bool // if true, this field may not be set using POST or PATCH requests
}

type ResourceFieldType int

const (
	PrimaryField   ResourceFieldType = iota + 1
	AttributeField
	RelationField
)

func (m *ResourceFields) GetPrimaryField() *ResourceField {
	for _, field := range *m {
		if field.Type == PrimaryField {
			return field
		}
	}

	return nil
}

func (m *ResourceFields) GetRelationFields() []*ResourceField {
	fields := make([]*ResourceField, 0)
	for _, field := range *m {
		if field.Type == RelationField {
			fields = append(fields, field)
		}
	}

	return fields
}
