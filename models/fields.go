package models

import (
	"reflect"
	"github.com/go-pg/pg/orm"
)

type ModelFields map[string]*ModelField

type ModelField struct {
	StructField *reflect.StructField
	Name        string         // jsonapi attribute name
	Type        FieldType      // jsonapi field type
	PGField     *orm.Field     // go-pg field
	Settings    *FieldSettings // jargo field settings
}

type FieldSettings struct {
	AllowFiltering bool // if true, filtering by this field is allowed
	AllowSorting   bool // if true, sorting by this field is allowed
}

type FieldType int

const (
	PrimaryField   FieldType = iota + 1
	AttributeField
	RelationField
)

func (m *ModelFields) GetPrimaryField() *ModelField {
	for _, field := range *m {
		if field.Type == PrimaryField {
			return field
		}
	}

	return nil
}

func (m *ModelFields) GetRelationFields() []*ModelField {
	fields := make([]*ModelField, 0)
	for _, field := range *m {
		if field.Type == RelationField {
			fields = append(fields, field)
		}
	}

	return fields
}
