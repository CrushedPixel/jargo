package internal

import (
	"github.com/crushedpixel/jargo/api"
	"reflect"
)

// resource registry
type Registry map[reflect.Type]*resource

func (r Registry) RegisterResource(resourceModelType reflect.Type) (resource api.Resource, err error) {
	// internally, jargo panics when parsing an invalid resource.
	// to be more gracious to the user, we recover from those
	// and return them as an error value.
	defer func() {
		if r := recover(); r != nil {
			resource = nil

			switch x := r.(type) {
			case error:
				err = x
			default:
				panic(r)
			}
		}
	}()

	var ok bool
	if resource, ok = r[resourceModelType]; ok {
		return
	}

	r.registerResource(resourceModelType)
	resource = r[resourceModelType]
	return
}

func (r Registry) registerResource(resourceModelType reflect.Type) {
	// check if schema is already registered
	// or currently being registered
	if _, ok := r[resourceModelType]; ok {
		return
	}

	// first, create schema definition including
	// joinJsonapiModel and joinPGModel
	schema := r.newSchemaDefinition(resourceModelType)

	// set value so the jsonapi and pg join models
	// can be accessed in r.generateSchemaModels
	r[resourceModelType] = &resource{schema}

	// create full schema definition including relations
	r.generateSchemaModels(schema)
}
