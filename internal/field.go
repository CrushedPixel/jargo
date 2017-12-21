package internal

import "reflect"

type resourceField struct {
	definition *fieldDefinition

	jsonapiFields []reflect.StructField
	pgFields      []reflect.StructField
}

func newResourceField(definition *fieldDefinition, registry *Registry) *resourceField {
	return &resourceField{
		definition: definition,
		jsonapiFields: generateJsonapiFields(definition, registry),
		pgFields: generatePGFields(definition, registry),
	}
}
