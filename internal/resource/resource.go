package resource

import "reflect"

type Resource struct {
	initialized bool
	definition  *resourceDefinition

	jsonapiModel reflect.Type
	pgModel      reflect.Type
}
