package jargo

import (
	"github.com/go-pg/pg/orm"
	"github.com/go-pg/pg/types"
	"reflect"
	"errors"
	"strings"
	"github.com/google/jsonapi"
	"io"
)

const (
	jsonApiIdField = "id"

	annotationJSONAPI   = "jsonapi"
	annotationPrimary   = "primary"
	annotationAttribute = "attr"
	annotationRelation  = "relation"

	annotationSeparator = ","

	annotationJargo  = "jargo"
	annotationFilter = "filter"
	annotationSort   = "sort"
)

var ErrInvalidModelType = errors.New("controller model must be pointer to struct")
var ErrInvalidIdField = errors.New("non-integer id fields are not supported")
var ErrClientId = errors.New("client-generated ids are not supported")

type Model struct {
	typ    reflect.Type // the database model type (pointer to a struct)
	Table  *orm.Table   // the database table
	Fields ModelFields  // information about the model's fields
}

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

func newModel(model interface{}) (*Model, error) {
	val := reflect.ValueOf(model)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return nil, ErrInvalidModelType
	}

	prtType := reflect.TypeOf(model)

	// get type of struct type points at
	typ := prtType.Elem()
	table := orm.Tables.Get(typ)

	fields := make(ModelFields)
	for i := 0; i < typ.NumField(); i++ {
		// iterate over model's fields,
		// parsing field information

		field := typ.Field(i)
		prop := getJsonApiProperties(&field)
		if prop == nil {
			continue
		}

		if prop.Type == PrimaryField && field.Type.Kind() != reflect.Int {
			return nil, ErrInvalidIdField
		}

		column := getSQLColumnName(table, &field)
		if column == nil {
			continue
		}

		settings := getFieldSettings(&field)
		if settings == nil {
			continue
		}

		fields[prop.Name] = &ModelField{
			StructField:       &field,
			Column:            *column,
			Settings:          settings,
			JsonApiProperties: prop,
		}
	}

	return &Model{
		typ:    typ,
		Table:  table,
		Fields: fields,
	}, nil
}

// returns a field's jsonapi name by parsing the jsonapi struct tag
func getJsonApiProperties(field *reflect.StructField) *JsonApiProperties {
	// parse jsonapi tag
	jsonTag := field.Tag.Get(annotationJSONAPI)
	if jsonTag == "" {
		return nil
	}

	args := strings.Split(jsonTag, annotationSeparator)
	if args[0] == annotationPrimary {
		return &JsonApiProperties{
			Name: jsonApiIdField,
			Type: PrimaryField,
		}
	} else if len(args) > 1 {
		name := args[1]
		if args[0] == annotationAttribute {
			return &JsonApiProperties{
				Name: name,
				Type: AttributeField,
			}
		} else if args[1] == annotationRelation {
			return &JsonApiProperties{
				Name: name,
				Type: RelationField,
			}
		}
	}

	return nil
}

// returns a field's SQL column name
func getSQLColumnName(table *orm.Table, field *reflect.StructField) *types.Q {
	for _, f := range table.Fields {
		if f.GoName == field.Name {
			return &f.Column
		}
	}

	return nil
}

// parses a field's jargo struct tags
func getFieldSettings(field *reflect.StructField) *FieldSettings {
	fieldSettings := new(FieldSettings)

	val, ok := field.Tag.Lookup(annotationJargo)
	if !ok {
		fieldSettings.AllowFiltering = true
		fieldSettings.AllowSorting = true
	} else {
		spl := strings.Split(val, annotationSeparator)
		for _, s := range spl {
			switch s {
			case annotationFilter:
				fieldSettings.AllowFiltering = true
			case annotationSort:
				fieldSettings.AllowSorting = true
			default:
				// TODO: error handling?
			}
		}
	}

	return fieldSettings
}

func (m *ModelFields) GetPrimaryField() *ModelField {
	for _, v := range *m {
		if v.JsonApiProperties.Type == PrimaryField {
			return v
		}
	}

	return nil
}

// returns a struct pointer
func (m *Model) NewInstance() interface{} {
	return reflect.New(m.typ).Interface()
}

// returns a pointer to a slice of struct pointers
func (m *Model) NewSlice() interface{} {
	return reflect.New(reflect.SliceOf(reflect.PtrTo(m.typ))).Interface()
}

func MarshalSlice(value interface{}) (jsonapi.Payloader, error) {
	// jsonapi.Marshal requires a slice instead of a pointer to a slice,
	// so we dereference it using reflection
	return jsonapi.Marshal(reflect.Indirect(reflect.ValueOf(value)).Interface())
}

// parses a jsonapi payload as sent in a POST request
func (m *Model) UnmarshalCreate(in io.Reader) (interface{}, error) {
	instance := m.NewInstance()
	err := jsonapi.UnmarshalPayload(in, instance)
	if err != nil {
		return nil, err
	}

	// disallow client-generated ids
	if reflect.ValueOf(instance).Elem().FieldByIndex(m.Fields.GetPrimaryField().StructField.Index).Int() != 0 {
		return nil, ErrClientId
	}

	return instance, nil
}

// parses a jsonapi payload as sent in a PATCH request,
// applying it to the existing entry, modifying its non-readonly values.
func (m *Model) UnmarshalUpdate(in io.Reader, instance interface{}) (interface{}, error) {
	err := jsonapi.UnmarshalPayload(in, instance)
	if err != nil {
		return nil, err
	}

	return instance, nil
}
