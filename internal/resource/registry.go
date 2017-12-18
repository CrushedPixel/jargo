package resource

import (
	"reflect"
	"github.com/pkg/errors"
	"fmt"
)

type Registry struct {
	resources map[reflect.Type]*Resource
}

// registers a resource. this resource can't be used before it's initialized.
func (r *Registry) RegisterResource(model interface{}) (*Resource, error) {
	return r.getResource(reflect.TypeOf(model))
}

func NewRegistry() *Registry {
	return &Registry{
		resources: make(map[reflect.Type]*Resource),
	}
}

// returns underlying struct type of struct pointer or slice of struct pointers
func getStructType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t
}

func (r *Registry) getResource(t reflect.Type) (*Resource, error) {
	// if resource is already registered, return resource
	if res, ok := r.resources[t]; ok {
		return res, nil
	}

	println(fmt.Sprintf("registering resource for type %v", t))

	definition, err := parseResourceType(t)
	if err != nil {
		return nil, err
	}

	res := &Resource{
		initialized: false,
		definition:  definition,
	}

	r.resources[t] = res
	return res, nil
}

func (r *Registry) InitializeResources() error {
	// resolve relationships
	for _, res := range r.resources {
		for _, field := range res.definition.fields {
			switch field.typ {
			case has, belongsTo, many2many:
				// register relation type as resource if it doesn't exist yet
				_, err := r.getResource(getStructType(field.structField.Type))
				if err != nil {
					return errors.New(fmt.Sprintf("error registering related resource: %s", err.Error()))
				}
			}
		}
	}

	// initialize resources
	for _, res := range r.resources {
		res.jsonapiModel = generateJsonapiStructType(res.definition)
		res.pgModel = generatePGModel(r, res.definition)

		res.initialized = true
	}

	return nil
}
