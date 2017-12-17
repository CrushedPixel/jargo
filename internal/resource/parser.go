package resource

import (
	"errors"
	"reflect"
	"crushedpixel.net/jargo/internal/parser"
	"github.com/iancoleman/strcase"
	"regexp"
)

const (
	jargoFieldTag   = "jargo"
	tableNameOption = "table"
)

var sqlNameRegex = regexp.MustCompile(`^[0-9a-zA-Z$_]+$`)
var memberNameRegex = regexp.MustCompile(`^[[:alnum:]]([a-zA-Z0-9\-_]*[[:alnum:]])?$`)

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
	idField, ok := t.FieldByName("Id")
	if !ok {
		return nil, errMissingIdField
	}

	if idField.Type != reflect.TypeOf(int64(0)) {
		return nil, errInvalidIdType
	}

	idTag, ok := idField.Tag.Lookup(jargoFieldTag)
	if !ok {
		return nil, errUnannotatedIdField
	}

	rd := new(resourceDefinition)

	defaultName := pluralize(strcase.ToSnake(t.Name()))
	parsed := parser.ParseJargoTagDefaultName(idTag, defaultName)

	rd.name = parsed.Name
	if table := parsed.Options[tableNameOption]; table != "" {
		rd.table = table
	} else {
		rd.table = rd.name
	}

	// validate member name
	if !isValidJsonapiMemberName(rd.name) {
		return nil, errInvalidMemberName
	}

	// validate table name
	if !isValidSQLName(rd.table) {
		return nil, errInvalidTableName
	}

	return rd, nil
}

func pluralize(val string) string {
	l := len(val)
	if l == 0 || val[l - 1] == 's' {
		return val
	}

	return val + "s"
}

func isValidJsonapiMemberName(val string) bool {
	return memberNameRegex.MatchString(val)
}

func isValidSQLName(val string) bool {
	return sqlNameRegex.MatchString(val)
}