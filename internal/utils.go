package internal

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

var sqlNameRegex = regexp.MustCompile(`^[0-9a-zA-Z$_]+$`)
var memberNameRegex = regexp.MustCompile(`^[[:alnum:]]([a-zA-Z0-9\-_]*[[:alnum:]])?$`)

func isValidJsonapiMemberName(val string) bool {
	return memberNameRegex.MatchString(val)
}

func isValidSQLName(val string) bool {
	return sqlNameRegex.MatchString(val)
}

func parseBoolOption(val string) bool {
	if val == "" {
		return true
	}

	b, err := strconv.ParseBool(val)
	if err != nil {
		panic(errors.New(fmt.Sprintf("error parsing bool option: %s", err.Error())))
	}
	return b
}

// wrapper types for at least a bit of type-safety when working with reflection.
type (
	resourceModelInstance struct {
		schema *Schema
		value  *reflect.Value // struct pointer value
	}
	jsonapiModelInstance struct {
		schema *Schema
		value  *reflect.Value // struct pointer value
	}
	pgModelInstance struct {
		schema *Schema
		value  *reflect.Value // struct pointer value
	}
	joinJsonapiModelInstance struct {
		schema *Schema
		value  *reflect.Value // struct pointer value
	}
	joinPGModelInstance struct {
		schema *Schema
		value  *reflect.Value // struct pointer value
	}
)
