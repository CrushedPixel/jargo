package jargo

import (
	"github.com/go-pg/pg/orm"
	"github.com/go-pg/pg/types"
	"reflect"
	"errors"
	"strings"
	"github.com/google/jsonapi"
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

var ErrInvalidModel = errors.New("controller model must be pointer to struct")

type Model struct {
	typ    *reflect.Type // the database model type (pointer to a struct)
	Table  *orm.Table    // the database table
	Fields ModelFields   // information about the model's fields
}

type ModelFields map[string]*ModelField

type ModelField struct {
	Column   types.Q // database column name
	Settings *FieldSettings
}

type FieldSettings struct {
	Filter bool // whether filtering by this field is allowed
	Sort   bool // whether sorting by this field is allowed
}

func newModel(model interface{}) (*Model, error) {
	val := reflect.ValueOf(model)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return nil, ErrInvalidModel
	}

	typ := reflect.TypeOf(model)

	// get type of struct type points at
	elem := typ.Elem()
	table := orm.Tables.Get(elem)

	fields := make(ModelFields)
	for i := 0; i < elem.NumField(); i++ {
		// iterate over model's fields,
		// parsing field information

		field := elem.Field(i)
		name := getJsonApiAttributeName(&field)
		if name == "" {
			continue
		}

		column := getSQLColumnName(table, &field)
		if column == nil {
			continue
		}

		settings := getFieldSettings(&field)
		if settings == nil {
			continue
		}

		fields[name] = &ModelField{
			Column:   *column,
			Settings: settings,
		}
	}

	return &Model{
		typ:    &typ,
		Table:  table,
		Fields: fields,
	}, nil
}

// returns a field's jsonapi name by parsing the jsonapi struct tag
func getJsonApiAttributeName(field *reflect.StructField) string {
	// parse jsonapi tag
	jsonTag := field.Tag.Get(annotationJSONAPI)
	if jsonTag == "" {
		return ""
	}

	args := strings.Split(jsonTag, annotationSeparator)
	if args[0] == annotationPrimary {
		return jsonApiIdField
	} else if len(args) > 1 &&
		(args[0] == annotationAttribute ||
			args[0] == annotationRelation) {
		return args[1]
	}

	return ""
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
	var filter, sort bool

	val, ok := field.Tag.Lookup(annotationJargo)
	if !ok {
		filter = true
		sort = true
	} else {
		spl := strings.Split(val, annotationSeparator)
		for _, s := range spl {
			switch s {
			case annotationFilter:
				filter = true
			case annotationSort:
				sort = true
			default:
				// TODO: error handling?
			}
		}
	}

	return &FieldSettings{
		Filter: filter,
		Sort:   sort,
	}
}

// returns a struct pointer
func (m *Model) NewInstance() interface{} {
	return reflect.New(*m.typ).Interface()
}

// returns a pointer to a slice of struct pointers
func (m *Model) NewSlice() interface{} {
	return reflect.New(reflect.SliceOf(*m.typ)).Interface()
}

func MarshalSlice(value interface{}) (jsonapi.Payloader, error) {
	// jsonapi.Marshal requires a slice instead of a pointer to a slice,
	// so we dereference it using reflection
	return jsonapi.Marshal(reflect.Indirect(reflect.ValueOf(value)).Interface())
}
