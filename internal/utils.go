package internal

import (
	"encoding"
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

// IdToString converts an id value into its string representation.
func IdToString(id interface{}) string {
	var str string

	switch i := id.(type) {
	case string:
		str = id.(string)
	case int, int8, int16, int32, int64:
		str = strconv.FormatInt(reflect.ValueOf(id).Int(), 10)
	case uint, uint8, uint16, uint32, uint64:
		str = strconv.FormatUint(reflect.ValueOf(id).Uint(), 10)
	case encoding.TextMarshaler:
		b, err := i.MarshalText()
		if err != nil {
			panic(err)
		}
		str = string(b)
	default:
		panic("invalid id type")
	}

	return str
}

// StringToId converts the string representation of an id
// into the target type.
func StringToId(id string, typ reflect.Type) interface{} {
	var val interface{}

	switch reflect.New(typ).Elem().Interface().(type) {
	case string:
		val = id
	case int, int8, int16, int32, int64:
		var err error
		val, err = strconv.ParseInt(id, 10, 0)
		if err != nil {
			panic(err)
		}
	case uint, uint8, uint16, uint32, uint64:
		var err error
		val, err = strconv.ParseUint(id, 10, 0)
		if err != nil {
			panic(err)
		}
	case encoding.TextMarshaler:
		// unmarshal value from string
		val = reflect.New(typ).Elem().Interface()
		val.(encoding.TextUnmarshaler).UnmarshalText([]byte(id))
	default:
		panic("invalid id type")
	}

	return val
}
