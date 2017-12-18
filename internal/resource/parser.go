package resource

import (
	"errors"
	"reflect"
	"crushedpixel.net/jargo/internal/parser"
	"github.com/iancoleman/strcase"
)

const (
	primaryFieldName  = "Id"
	primaryColumnName = "id"

	unexportedFieldName = "-"

	jargoFieldTag   = "jargo"
	tableNameOption = "table"

	columnNameOption = "column"
)

type resourceDefinition struct {
	fields []*field
	name   string
	table  string
}

var errInvalidModelType = errors.New("model has to be struct")
var errMissingIdField = errors.New("missing id field")
var errInvalidIdType = errors.New("id field must be of type int64")
var errUnannotatedIdField = errors.New("id field is missing jargo annotation")
var errInvalidMemberName = errors.New(`member name has to adhere to the jsonapi specification and not include characters marked as "not recommended"`)
var errInvalidTableName = errors.New("table name may only consist of [0-9,a-z,A-Z$_]")
var errInvalidColumnName = errors.New("column name may only consist of [0-9,a-z,A-Z$_]")

func parseResourceStruct(model interface{}) (*resourceDefinition, error) {
	t := reflect.TypeOf(model)
	if t.Kind() != reflect.Struct {
		return nil, errInvalidModelType
	}

	// parse Id field
	name, table, err := parseIdField(t)
	if err != nil {
		return nil, err
	}

	rd := new(resourceDefinition)
	rd.name = name
	rd.table = table

	// parse all attributes
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		field, err := parseAttrField(&f)
		if err != nil {
			return nil, err
		}

		rd.fields = append(rd.fields, field)
	}

	return rd, nil
}

func parseIdField(t reflect.Type) (name string, table string, err error) {
	idField, ok := t.FieldByName(primaryFieldName)
	if !ok {
		return "", "", errMissingIdField
	}

	if idField.Type != reflect.TypeOf(int64(0)) {
		return "", "", errInvalidIdType
	}

	idTag, ok := idField.Tag.Lookup(jargoFieldTag)
	if !ok {
		return "", "", errUnannotatedIdField
	}

	defaultName := pluralize(strcase.ToSnake(t.Name()))
	parsed := parser.ParseJargoTagDefaultName(idTag, defaultName)

	name = parsed.Name
	if tbl := parsed.Options[tableNameOption]; tbl != "" {
		table = tbl
	} else {
		table = name
	}

	// validate member name
	if !isValidJsonapiMemberName(name) {
		return "", "", errInvalidMemberName
	}

	// validate table name
	if !isValidSQLName(table) {
		return "", "", errInvalidTableName
	}

	return name, table, nil
}

func parseAttrField(f *reflect.StructField) (*field, error) {
	field := new(field)

	if f.Name == primaryFieldName {
		field.typ = id
		field.column = primaryColumnName
		return field, nil
	}

	defaultName := strcase.ToSnake(f.Name)
	parsed := parser.ParseJargoTagDefaultName(f.Tag.Get(jargoFieldTag), defaultName)

	field.name = parsed.Name
	if column := parsed.Options[columnNameOption]; column != "" {
		field.column = column
	} else {
		if field.name != unexportedFieldName {
			field.column = field.name
		} else {
			field.column = defaultName
		}
	}

	// validate member name
	if field.name != unexportedFieldName && !isValidJsonapiMemberName(field.name) {
		return nil, errInvalidMemberName
	}

	// validate table name
	if !isValidSQLName(field.column) {
		return nil, errInvalidColumnName
	}

	// TODO parse rest of field attributes
	return field, nil
}
