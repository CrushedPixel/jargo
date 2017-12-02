package models

import (
	"github.com/go-pg/pg/types"
	"reflect"
	"errors"
	"strings"
	"github.com/google/jsonapi"
	"io"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
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
var ErrInvalidIdType = errors.New("non-integer id fields are not supported")
var ErrMissingSQLPK = errors.New("jsonapi primary field must be marked as primary key in sql struct tag")
var ErrClientId = errors.New("client-generated ids are not supported")

type Model struct {
	typ    reflect.Type // the database model type (pointer to a struct)
	Table  *orm.Table   // the database table
	Fields ModelFields  // information about the model's fields
}

func New(model interface{}) (*Model, error) {
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

		if prop.Type == PrimaryField {
			if field.Type.Kind() != reflect.Int {
				return nil, ErrInvalidIdType
			}

			// ensure the primary key field has the sql pk struct tag
			sqlTag := field.Tag.Get("sql")
			spl := strings.Split(sqlTag, annotationSeparator)
			if len(spl) < 2 || spl[1] != "pk" {
				return nil, ErrMissingSQLPK
			}
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

func (m *Model) Select(db *pg.DB) *Query {
	instance := m.newSlice()
	return &Query{
		Query: db.Model(instance),
		Type:  Select,
		value: reflect.ValueOf(instance),
	}
}

func (m *Model) SelectOne(db *pg.DB) *Query {
	instance := m.newInstance()
	return &Query{
		Query: db.Model(instance),
		Type:  Select,
		value: reflect.ValueOf(instance),
	}
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

// returns a struct pointer
func (m *Model) newInstance() interface{} {
	return reflect.New(m.typ).Interface()
}

// returns a pointer to a slice of struct pointers
func (m *Model) newSlice() interface{} {
	return reflect.New(reflect.SliceOf(reflect.PtrTo(m.typ))).Interface()
}

func Marshal(value interface{}) (jsonapi.Payloader, error) {
	t := reflect.TypeOf(value)
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Slice {
		// jsonapi.Marshal requires a slice instead of a pointer to a slice,
		// so we dereference it using reflection
		return jsonapi.Marshal(reflect.Indirect(reflect.ValueOf(value)).Interface())
	}

	return jsonapi.Marshal(value)
}

// creates the database table for the model if it doesn't exist
func (m *Model) CreateTable(db *pg.DB) error {
	return db.CreateTable(m.newInstance(), &orm.CreateTableOptions{IfNotExists: true})
}

// parses a jsonapi payload as sent in a POST request
func (m *Model) UnmarshalCreate(in io.Reader) (interface{}, error) {
	instance := m.newInstance()
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
