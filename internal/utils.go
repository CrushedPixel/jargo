package internal

import (
	"encoding"
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

// isSet returns whether an option is set.
// It panics if a value is specified.
func isSet(options map[string]string, key string) bool {
	val, ok := options[key]
	if val != "" {
		panic(fmt.Errorf(`option "%s" does not accept a value`, key))
	}
	return ok
}

// moreThanOneTrue returns whether more than one
// of the boolean values passed are true.
func moreThanOneTrue(bools ...bool) bool {
	var found bool
	for _, b := range bools {
		if b {
			if found {
				return true
			}
			found = true
		}
	}
	return false
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
	typ := reflect.TypeOf(id)
	if isTextMarshaler(typ) {
		b, err := id.(encoding.TextMarshaler).MarshalText()
		if err != nil {
			panic(err)
		}
		return string(b)
	} else if pointerTypeIsTextMarshaler(typ) {
		i := reflect.New(typ)
		i.Elem().Set(reflect.ValueOf(id))

		b, err := i.Interface().(encoding.TextMarshaler).MarshalText()
		if err != nil {
			panic(err)
		}
		return string(b)
	}

	switch id.(type) {
	case string:
		return id.(string)
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(id).Int(), 10)
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(reflect.ValueOf(id).Uint(), 10)
	default:
		panic("invalid id type")
	}
}

// StringToId converts the string representation of an id
// into the target type.
func StringToId(id string, typ reflect.Type) interface{} {
	if isTextUnmarshaler(typ) || pointerTypeIsTextUnmarshaler(typ) {
		// unmarshal value from string
		return StringToTextUnmarshaler(id, typ)
	}

	switch reflect.New(typ).Elem().Interface().(type) {
	case string:
		return id
	case int, int8, int16, int32, int64:
		val, err := strconv.ParseInt(id, 10, 0)
		if err != nil {
			panic(err)
		}
		return val
	case uint, uint8, uint16, uint32, uint64:
		val, err := strconv.ParseUint(id, 10, 0)
		if err != nil {
			panic(err)
		}
		return val
	default:
		panic("invalid id type")
	}
}

// StringToTextUnmarshaler converts a string to an instance
// of the given type, whose pointer type or itself
// must implement encoding.TextMarshaler.
func StringToTextUnmarshaler(id string, typ reflect.Type) interface{} {
	// create a new instance of the unmarshaler type
	val := reflect.New(typ).Elem().Interface()

	// sometimes, only the pointer type of a type
	// implements an interface (e.g. *uuid.UUID),
	// so we have to check whether the type we have
	// does actually implement the interface,
	// or whether whe have to operate on the pointer.

	if isTextUnmarshaler(typ) {
		if err := val.(encoding.TextUnmarshaler).UnmarshalText([]byte(id)); err != nil {
			panic(err)
		}
	} else if pointerTypeIsTextUnmarshaler(typ) {
		val = reflect.New(typ).Interface()
		if err := val.(encoding.TextUnmarshaler).UnmarshalText([]byte(id)); err != nil {
			panic(err)
		}
		val = reflect.ValueOf(val).Elem().Interface()
	} else {
		panic("typ is not an encoding.TextUnmarshaler")
	}

	return val
}
