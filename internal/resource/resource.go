package resource

import (
	"reflect"
	"github.com/go-pg/pg"
)

// to be used whenever a variable that is only available
// after initialization is accessed in a function
var errNotInitialized = "resource has not been initialized yet"

type Resource struct {
	Type reflect.Type // type of the resource model

	definition  *resourceDefinition
	initialized bool

	// fields only available after initialization by Registry
	fields []*resourceField

	pgModel             reflect.Type
	staticJsonapiFields []reflect.StructField
}

func (r *Resource) jsonapiModel(fs *FieldSet) reflect.Type {
	if !r.initialized {
		panic(errNotInitialized)
	}
	if fs.resource != r {
		panic("trying to generate jsonapi model from field set of different resource")
	}

	fields := append(r.staticJsonapiFields, fs.jsonapiFields()...)
	return reflect.StructOf(fields)
}

// returns a struct pointer
func (r *Resource) newModelInstance() interface{} {
	if !r.initialized {
		panic(errNotInitialized)
	}
	return reflect.New(r.pgModel).Interface()
}

// returns a pointer to a slice of struct pointers
func (r *Resource) newModelSlice() interface{} {
	if !r.initialized {
		panic(errNotInitialized)
	}
	return reflect.New(reflect.SliceOf(reflect.PtrTo(r.pgModel))).Interface()
}

func (r *Resource) Select(db *pg.DB) *Query {
	return newQuery(db, r, typeSelect, true)
}

func (r *Resource) SelectOne(db *pg.DB) *Query {
	return newQuery(db, r, typeSelect, false)
}

func (r *Resource) SelectById(db *pg.DB, id interface{}) *Query {
	q := r.SelectOne(db)
	q.Where("Id = ?", id)
	return q
}
