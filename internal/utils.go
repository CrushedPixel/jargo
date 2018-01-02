package internal

import (
	"regexp"
	"reflect"
	"strconv"
	"crushedpixel.net/jargo/api"
	"github.com/go-pg/pg/types"
)

var sqlNameRegex = regexp.MustCompile(`^[0-9a-zA-Z$_]+$`)
var memberNameRegex = regexp.MustCompile(`^[[:alnum:]]([a-zA-Z0-9\-_]*[[:alnum:]])?$`)

func isValidJsonapiMemberName(val string) bool {
	return memberNameRegex.MatchString(val)
}

func isValidSQLName(val string) bool {
	return sqlNameRegex.MatchString(val)
}

func parseBoolOption(val string) (bool, error) {
	if val == "" {
		return true, nil
	}

	return strconv.ParseBool(val)
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

func resourceModelToPGModel(schema api.Schema, resourceModelInstance interface{}) interface{} {
	schemaInstance, err := schema.ParseResourceModel(resourceModelInstance)
	if err != nil {
		panic(err)
	}
	pgModelInstance, err := schemaInstance.ToPGModel()
	if err != nil {
		panic(err)
	}
	return pgModelInstance
}

func resourceModelToJsonapiModel(schema api.Schema, resourceModelInstance interface{}) interface{} {
	schemaInstance, err := schema.ParseResourceModel(resourceModelInstance)
	if err != nil {
		panic(err)
	}
	jsonapiModelInstance, err := schemaInstance.ToJsonapiModel()
	if err != nil {
		panic(err)
	}
	return jsonapiModelInstance
}

func pgModelToResourceModel(schema api.Schema, pgModelInstance interface{}) interface{} {
	schemaInstance, err := schema.ParsePGModel(pgModelInstance)
	if err != nil {
		panic(err)
	}
	resourceModelInstance, err := schemaInstance.ToResourceModel()
	if err != nil {
		panic(err)
	}
	return resourceModelInstance
}

func jsonapiModelToResourceModel(schema api.Schema, pgModelInstance interface{}) interface{} {
	schemaInstance, err := schema.ParseJsonapiModel(pgModelInstance)
	if err != nil {
		panic(err)
	}
	resourceModelInstance, err := schemaInstance.ToResourceModel()
	if err != nil {
		panic(err)
	}
	return resourceModelInstance
}

// escapes a go-pg column string according to postgres rules.
// example: user.id => "user"."id"
func escapePGField(field string) string {
	var b []byte
	b = types.AppendField(b, field, 1)
	return string(b)
}
