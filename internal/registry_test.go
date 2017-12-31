package internal

import (
	"testing"
	"reflect"
	"github.com/stretchr/testify/assert"
	"github.com/google/jsonapi"
	"os"
	"encoding/json"
)

type Age int

type Basic struct {
	Id   int64 `jargo:""`
	Name string
	Age  int
}

type CustomType struct {
	Id   int64 `jargo:""`
	Name string
	Age  Age
}

type HasOne struct {
	Id    int64  `jargo:""`
	Name  string
	Basic *Basic `jargo:",has"`
}

type HasMany struct {
	Id    int64    `jargo:""`
	Name  string
	Basic []*Basic `jargo:",has"`
}

type HasSelf struct {
	Id   int64    `jargo:""`
	Name string
	Self *HasSelf `jargo:",has"`
}

type CircularHas struct {
	Id   int64              `jargo:""`
	Name string
	B    *CircularBelongsTo `jargo:",has"`
}

type CircularBelongsTo struct {
	Id   int64        `jargo:""`
	Name string
	A    *CircularHas `jargo:",belongsTo"`
}

func TestRegistry_RegisterSchema(t *testing.T) {
	r := make(Registry)
	// create schema from struct type
	s, err := r.RegisterResource(reflect.TypeOf(Basic{}))
	assert.Nil(t, err)
	// TODO: test field generation

	// parse resource model instance to schema instance
	b := Basic{
		Id:   1,
		Name: "Peter",
		Age:  50,
	}
	instance, err := s.ParseResourceModel(&b)
	assert.Nil(t, err)

	// convert schema instance back to resource model instance
	b1, err := instance.ToResourceModel()
	assert.Nil(t, err)

	j1, err := instance.ToJsonapiModel()
	assert.Nil(t, err)
	jsonapi.MarshalPayloadWithoutIncluded(os.Stdout, j1)

	// expect original and converted instances to be equal
	assert.Equal(t, &b, b1)

	// test conversion to jsonapi model
	_, err = instance.ToJsonapiModel()
	assert.Nil(t, err)

	// test conversion to pg model
	_, err = instance.ToPGModel()
	assert.Nil(t, err)

	// test handling of custom types
	s, err = r.RegisterResource(reflect.TypeOf(CustomType{}))
	assert.EqualError(t, err, errInvalidAttrFieldType(reflect.TypeOf(Age(0))).Error())

	// test hasOne relations
	s, err = r.RegisterResource(reflect.TypeOf(HasOne{}))
	assert.Nil(t, err)

	h := HasOne{
		Id:    1,
		Name:  "Peter",
		Basic: &b,
	}
	instance, err = s.ParseResourceModel(&h)
	assert.Nil(t, err)

	j1, err = instance.ToJsonapiModel()
	assert.Nil(t, err)
	jsonapi.MarshalPayloadWithoutIncluded(os.Stdout, j1)

	// test hasMany relations
	s, err = r.RegisterResource(reflect.TypeOf(HasMany{}))
	assert.Nil(t, err)

	hm := HasMany{
		Id:    1,
		Name:  "Peter",
		Basic: []*Basic{&b, b1.(*Basic)},
	}
	instance, err = s.ParseResourceModel(&hm)
	assert.Nil(t, err)

	j1, err = instance.ToJsonapiModel()
	assert.Nil(t, err)
	jsonapi.MarshalPayloadWithoutIncluded(os.Stdout, j1)

	// test hasSelf relations
	s, err = r.RegisterResource(reflect.TypeOf(HasSelf{}))
	assert.Nil(t, err)

	hs := HasSelf{
		Id:   1,
		Name: "Peter",
		Self: nil,
	}
	instance, err = s.ParseResourceModel(&hs)
	assert.Nil(t, err)

	j1, err = instance.ToJsonapiModel()
	assert.Nil(t, err)
	jsonapi.MarshalPayloadWithoutIncluded(os.Stdout, j1)

	// test circular relations
	s, err = r.RegisterResource(reflect.TypeOf(CircularHas{}))
	assert.Nil(t, err)

	hb := CircularBelongsTo{
		Id:   2,
		Name: "Pan",
		A:    nil,
	}
	ha := CircularHas{
		Id:   1,
		Name: "Peter",
		B:    &hb,
	}
	instance, err = s.ParseResourceModel(&ha)
	assert.Nil(t, err)

	j1, err = instance.ToJsonapiModel()
	assert.Nil(t, err)
	jsonapi.MarshalPayloadWithoutIncluded(os.Stdout, j1)

	p1, err := instance.ToResourceModel()
	assert.Nil(t, err)
	bytes, err := json.Marshal(p1)
	assert.Nil(t, err)
	println(string(bytes))

	// more circular relations testing
	hb = CircularBelongsTo{
		Id:   2,
		Name: "Pan",
		A:    &ha,
	}
	ha = CircularHas{
		Id:   1,
		Name: "Peter",
		B:    &hb,
	}
	instance, err = s.ParseResourceModel(&ha)
	assert.Nil(t, err)

	j1, err = instance.ToJsonapiModel()
	assert.Nil(t, err)
	jsonapi.MarshalPayloadWithoutIncluded(os.Stdout, j1)

	p1, err = instance.ToPGModel()
	assert.Nil(t, err)
	bytes, err = json.Marshal(p1)
	assert.Nil(t, err)
	println(string(bytes))
}
