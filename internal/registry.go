package internal

import (
	"reflect"
	"crushedpixel.net/jargo/api"
)

// resource registry
type Registry map[reflect.Type]*resource

func (r Registry) RegisterResource(resourceModelType reflect.Type) (api.Resource, error) {
	if resource, ok := r[resourceModelType]; ok {
		return resource, nil
	}

	r.registerResource(resourceModelType)
	return r[resourceModelType], nil
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
