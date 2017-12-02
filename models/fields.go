package models

import (
	"reflect"
	"github.com/go-pg/pg/types"
)

type ModelFields map[string]*ModelField

type ModelField struct {
	StructField       *reflect.StructField
	Column            types.Q // database column name
	Settings          *FieldSettings
	JsonApiProperties *JsonApiProperties
}

type FieldSettings struct {
	AllowFiltering bool // if true, filtering by this field is allowed
	AllowSorting   bool // if true, sorting by this field is allowed
}


type JsonApiFieldType int

const (
	PrimaryField   JsonApiFieldType = iota + 1
	AttributeField
	RelationField
)

type JsonApiProperties struct {
	Name string
	Type JsonApiFieldType
}

func (m *ModelFields) GetPrimaryField() *ModelField {
	for _, v := range *m {
		if v.JsonApiProperties.Type == PrimaryField {
			return v
		}
	}

	return nil
}