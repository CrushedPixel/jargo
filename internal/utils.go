package internal

import (
	"regexp"
	"reflect"
	"strconv"
	"github.com/go-pg/pg/types"
	"errors"
	"fmt"
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
		schema *schema
		value  *reflect.Value // struct pointer value
	}
	jsonapiModelInstance struct {
		schema *schema
		value  *reflect.Value // struct pointer value
	}
	pgModelInstance struct {
		schema *schema
		value  *reflect.Value // struct pointer value
	}
	joinJsonapiModelInstance struct {
		schema *schema
		value  *reflect.Value // struct pointer value
	}
	joinPGModelInstance struct {
		schema *schema
		value  *reflect.Value // struct pointer value
	}
)

// escapes a go-pg column string according to postgres rules.
// example: user.id => "user"."id"
func escapePGField(field string) string {
	var b []byte
	b = types.AppendField(b, field, 1)
	return string(b)
}
