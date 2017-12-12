package models

import (
	"reflect"
	"errors"
	"strings"
	"github.com/google/jsonapi"
	"io"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"fmt"
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
var ErrMissingPrimaryField = errors.New("missing jsonapi primary field")
var ErrMissingSQLPK = errors.New("jsonapi primary field must be marked as primary key in sql struct tag")
var ErrClientId = errors.New("client-generated ids are not supported")

type Model struct {
	Name   string      // resource name for jsonapi
	Table  *orm.Table  // the database table
	Fields ModelFields // information about the model's fields

	typ reflect.Type // the database model type (pointer to a struct)
}

// internally used for parsing jsonapi struct tags
type jsonapiField struct {
	Name string
	Type FieldType
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

	modelName := ""

	fields := make(ModelFields)
	for i := 0; i < typ.NumField(); i++ {
		// iterate over model's fields,
		// parsing field information

		field := typ.Field(i)
		prop := parseJsonapiTag(&field)
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

			modelName = prop.Name

			// the name of the field itself is "id"
			prop.Name = jsonApiIdField
		}

		var pgField *orm.Field
		// relationships are persisted with a non-jsonapi-annotated id field
		if prop.Type != RelationField {
			pgField = getPGField(table, &field)
			if pgField == nil {
				continue
			}
		} else {
			pgField = getPGField(table, &field)
			// TODO
			println(fmt.Sprintf("relationship pgfield %v", pgField))
		}

		settings := getFieldSettings(&field)
		if settings == nil {
			continue
		}

		fields[prop.Name] = &ModelField{
			StructField: &field,
			Name:        prop.Name,
			Type:        prop.Type,
			PGField:     pgField,
			Settings:    settings,
		}
	}

	if modelName == "" {
		return nil, ErrMissingPrimaryField
	}

	return &Model{
		Name:   modelName,
		Table:  table,
		Fields: fields,
		typ:    typ,
	}, nil
}

func (m *Model) selectAllColumns(q *orm.Query) {
	// select all columns ("table".*)
	q.Column(fmt.Sprintf("%s.*", m.Table.ModelName))

	// include all relationships
	for _, field := range m.Fields.GetRelationFields() {
		println(fmt.Sprintf("ey %v", field))
		q.Column(field.StructField.Name)
	}
}

func (m *Model) Select(db *pg.DB) *Query {
	instance := m.newSlice()

	q := db.Model(instance)
	m.selectAllColumns(q)

	return &Query{
		Query: q,
		Type:  Select,
		value: reflect.ValueOf(instance),
	}
}

func (m *Model) SelectOne(db *pg.DB) *Query {
	instance := m.newInstance()

	q := db.Model(instance)
	m.selectAllColumns(q)

	return &Query{
		Query: q,
		Type:  Select,
		value: reflect.ValueOf(instance),
	}
}

// returns a field's jsonapi name by parsing the jsonapi struct tag
func parseJsonapiTag(field *reflect.StructField) *jsonapiField {
	jsonTag := field.Tag.Get(annotationJSONAPI)
	if jsonTag == "" {
		return nil
	}

	args := strings.Split(jsonTag, annotationSeparator)
	if len(args) > 1 {
		name := args[1]
		var typ FieldType
		switch args[0] {
		case annotationPrimary:
			typ = PrimaryField
			break
		case annotationAttribute:
			typ = AttributeField
			break
		case annotationRelation:
			typ = RelationField
			break
		default:
			return nil
		}

		return &jsonapiField{
			Name: name,
			Type: typ,
		}
	}

	return nil
}

// returns a field's SQL column name
func getPGField(table *orm.Table, field *reflect.StructField) *orm.Field {
	for _, f := range table.Fields {
		if f.GoName == field.Name {
			return f
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
