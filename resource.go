package jargo

import (
	"reflect"
	"errors"
	"strings"
	"github.com/google/jsonapi"
	"io"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"fmt"
	"github.com/ulule/deepcopier"
	"strconv"
	"encoding/json"
	"bytes"
)

const (
	jsonApiIdField = "id"

	annotationJSONAPI   = "jsonapi"
	annotationPrimary   = "primary"
	annotationAttribute = "attr"
	annotationRelation  = "relation"

	annotationSeparator         = ","
	annotationKeyValueSeparator = ":"

	annotationJargo    = "jargo"
	annotationFilter   = "filter"
	annotationSort     = "sort"
	annotationReadonly = "readonly"
)

var ErrInvalidResourceType = errors.New("resource must be pointer to struct")
var ErrInvalidIdType = errors.New("only int64 id fields are supported")
var ErrMissingPrimaryField = errors.New("missing jsonapi primary field")
var ErrMissingSQLPK = errors.New("jsonapi primary field must be marked as primary key in sql struct tag")
var ErrClientId = errors.New("client-generated ids are not supported")
var ErrInvalidOriginal = errors.New("original must be a struct pointer")
var ErrMismatchedId = errors.New("payload id does not match original id")

type Resource struct {
	Name   string         // resource name parsed from jsonapi
	Table  *orm.Table     // the database table
	Fields ResourceFields // information about the model's fields

	typ reflect.Type // the database model type (pointer to a struct)
}

// internally used for parsing jsonapi struct tags
type jsonapiField struct {
	Name string
	Type ResourceFieldType
}

func NewResource(model interface{}) (*Resource, error) {
	val := reflect.ValueOf(model)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return nil, ErrInvalidResourceType
	}

	prtType := reflect.TypeOf(model)

	// get type of struct type points at
	typ := prtType.Elem()
	table := orm.Tables.Get(typ)

	modelName := ""

	fields := make(ResourceFields)
	for i := 0; i < typ.NumField(); i++ {
		// iterate over model's fields,
		// parsing field information

		field := typ.Field(i)
		prop := parseJsonapiTag(&field)
		if prop == nil {
			continue
		}

		if prop.Type == PrimaryField {
			if field.Type.Kind() != reflect.Int64 {
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

		settings, err := getFieldSettings(&field)
		if err != nil {
			return nil, err
		}

		fields[prop.Name] = &ResourceField{
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

	return &Resource{
		Name:   modelName,
		Table:  table,
		Fields: fields,
		typ:    typ,
	}, nil
}

func (m *Resource) selectAllColumns(q *Query) {
	// select all columns ("table".*)
	q.Column(fmt.Sprintf("%s.*", m.Table.ModelName))

	// include all relationships
	for _, field := range m.Fields.GetRelationFields() {
		q.Column(field.StructField.Name)
	}
}

func (m *Resource) Select(db *pg.DB) *Query {
	q := NewSelectQuery(db, m.newSlice())
	m.selectAllColumns(q)

	return q
}

func (m *Resource) SelectOne(db *pg.DB) *Query {
	q := NewSelectQuery(db, m.newInstance())
	m.selectAllColumns(q)

	return q
}

func (m *Resource) SelectById(db *pg.DB, id string) *Query {
	q := m.SelectOne(db)
	q.Where("id = ?", id)

	return q
}

func (m *Resource) DeleteById(db *pg.DB, id string) *Query {
	q := NewDeleteQuery(db, m.newInstance())
	q.Where("id = ?", id)

	return q
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
		var typ ResourceFieldType
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
func getFieldSettings(field *reflect.StructField) (*FieldSettings, error) {
	fieldSettings := new(FieldSettings)

	fieldSettings.AllowFiltering = true
	fieldSettings.AllowSorting = true
	fieldSettings.Readonly = false

	val, ok := field.Tag.Lookup(annotationJargo)
	if ok {
		spl := strings.Split(val, annotationSeparator)
		for _, s := range spl {
			kv := strings.Split(s, annotationKeyValueSeparator)
			if len(kv) != 2 {
				return nil, errors.New(fmt.Sprintf(`jargo option "%s" must be in "key:value" format`, s))
			}

			// parse boolean value
			val, err := strconv.ParseBool(kv[1])
			if err != nil {
				return nil, errors.New(fmt.Sprintf(`jargo option value: "%s". must be a boolean`, kv[1]))
			}

			switch kv[0] {
			case annotationFilter:
				fieldSettings.AllowFiltering = val
			case annotationSort:
				fieldSettings.AllowSorting = val
			case annotationReadonly:
				fieldSettings.Readonly = val
			default:
				return nil, errors.New(fmt.Sprintf(`unknown jargo option key: "%s"`, kv[0]))
			}
		}
	}

	return fieldSettings, nil
}

// returns a struct pointer
func (m *Resource) newInstance() interface{} {
	return reflect.New(m.typ).Interface()
}

// returns a pointer to a slice of struct pointers
func (m *Resource) newSlice() interface{} {
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
func (m *Resource) CreateTable(db *pg.DB) error {
	return db.CreateTable(m.newInstance(), &orm.CreateTableOptions{IfNotExists: true})
}

// parses a jsonapi payload as sent in a POST request
func (m *Resource) UnmarshalCreate(in io.Reader) (interface{}, error) {
	original := m.newInstance()

	instance := m.newInstance()
	err := jsonapi.UnmarshalPayload(in, instance)
	if err != nil {
		return nil, err
	}

	// disallow client-generated ids
	if reflect.ValueOf(instance).Elem().FieldByIndex(m.Fields.GetPrimaryField().StructField.Index).Int() != 0 {
		return nil, ErrClientId
	}

	// revert changes to all readonly fields
	m.revertReadonlyChanges(reflect.ValueOf(original).Elem(), reflect.ValueOf(instance).Elem())

	return instance, nil
}

// parses a jsonapi payload as sent in a PATCH request,
// applying it to the existing entry, modifying its non-readonly values.
func (m *Resource) UnmarshalUpdate(in io.Reader, original interface{}, originalId string) (interface{}, error) {
	originalValue := reflect.ValueOf(original)
	// original must be struct pointer
	if originalValue.Kind() != reflect.Ptr || originalValue.Elem().Kind() != reflect.Struct {
		return nil, ErrInvalidOriginal
	}

	originalStruct := originalValue.Elem().Interface()
	instance := reflect.New(originalValue.Type().Elem()).Interface()

	// copy original instance to be able to retain original values
	err := deepcopier.Copy(originalStruct).To(instance)
	if err != nil {
		return nil, err
	}

	// ensure id value of payload matches original id
	// read payload into buffer to be able to read it multiple times
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)

	// manually parse payload to be able to access id value
	payload := new(jsonapi.OnePayload)
	err = json.Unmarshal(buf.Bytes(), payload)
	if err != nil {
		return nil, err
	}

	if payload.Data.ID != originalId {
		return nil, ApiErrInvalidPayload(ErrMismatchedId)
	}

	// unmarshal payload into instance
	err = jsonapi.UnmarshalPayload(buf, instance)
	if err != nil {
		return nil, err
	}

	// revert changes to all readonly fields
	originalStructValue := reflect.ValueOf(originalStruct)
	instanceStructValue := reflect.ValueOf(instance).Elem()
	m.revertReadonlyChanges(originalStructValue, instanceStructValue)

	return instance, nil
}

func (m *Resource) revertReadonlyChanges(original reflect.Value, instance reflect.Value) {
	for _, field := range m.Fields {
		if field.Settings.Readonly {
			instance.FieldByIndex(field.StructField.Index).
				Set(original.FieldByIndex(field.StructField.Index))
		}
	}
}
