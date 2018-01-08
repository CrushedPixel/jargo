package internal

import (
	"reflect"
)

type SchemaRegistry map[reflect.Type]*Schema

func (r SchemaRegistry) RegisterSchema(modelType reflect.Type) (schema *Schema, err error) {
	// internally, jargo panics when parsing an invalid resource.
	// to be more gracious to the user, we recover from those
	// and return them as an error value.
	defer func() {
		if r := recover(); r != nil {
			schema = nil

			switch x := r.(type) {
			case error:
				err = x
			default:
				panic(r)
			}
		}
	}()

	var ok bool
	if schema, ok = r[modelType]; ok {
		return
	}

	r.registerSchema(modelType)
	schema = r[modelType]
	return
}

func (r SchemaRegistry) registerSchema(modelType reflect.Type) {
	// check if Schema is already registered
	// or currently being registered
	if _, ok := r[modelType]; ok {
		return
	}

	// first, create Schema definition including
	// joinJsonapiModel and joinPGModel
	schema := r.newSchemaDefinition(modelType)

	// set value so the jsonapi and pg join models
	// can be accessed in r.generateSchemaModels
	r[modelType] = schema

	// create full Schema definition including relations
	r.generateSchemaModels(schema)
}
