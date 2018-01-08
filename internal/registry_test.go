package internal

import (
	"encoding/json"
	"github.com/google/jsonapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"reflect"
	"testing"
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
	Id    int64 `jargo:""`
	Name  string
	Basic *Basic `jargo:",has"`
}

type HasMany struct {
	Id    int64 `jargo:""`
	Name  string
	Basic []*Basic `jargo:",has"`
}

type HasSelf struct {
	Id   int64 `jargo:""`
	Name string
	Self *HasSelf `jargo:",has"`
}

type CircularHas struct {
	Id   int64 `jargo:""`
	Name string
	B    *CircularBelongsTo `jargo:",has"`
}

type CircularBelongsTo struct {
	Id   int64 `jargo:""`
	Name string
	A    *CircularHas `jargo:",belongsTo"`
}

func TestRegistry_RegisterSchema(t *testing.T) {
	r := make(ResourceRegistry)
	// create schema from struct type
	s, err := r.RegisterResource(reflect.TypeOf(Basic{}))
	require.Nil(t, err)
	// TODO: test field generation

	// parse resource model instance to schema instance
	b := Basic{
		Id:   1,
		Name: "Peter",
		Age:  50,
	}
	instance := s.ParseResourceModel(&b)

	// convert schema instance back to resource model instance
	b1 := instance.ToResourceModel()

	j1 := instance.ToJsonapiModel()
	jsonapi.MarshalPayloadWithoutIncluded(os.Stdout, j1)

	// expect original and converted instances to be equal
	assert.Equal(t, &b, b1)

	// test conversion to jsonapi model
	_ = instance.ToJsonapiModel()

	// test conversion to pg model
	_ = instance.ToPGModel()

	// test handling of custom types
	s, err = r.RegisterResource(reflect.TypeOf(CustomType{}))
	assert.EqualError(t, err, errInvalidAttrFieldType(reflect.TypeOf(Age(0))).Error())

	// test hasOne relations
	s, err = r.RegisterResource(reflect.TypeOf(HasOne{}))
	require.Nil(t, err)

	h := HasOne{
		Id:    1,
		Name:  "Peter",
		Basic: &b,
	}
	instance = s.ParseResourceModel(&h)

	j1 = instance.ToJsonapiModel()
	jsonapi.MarshalPayloadWithoutIncluded(os.Stdout, j1)

	// test hasMany relations
	s, err = r.RegisterResource(reflect.TypeOf(HasMany{}))
	require.Nil(t, err)

	hm := HasMany{
		Id:    1,
		Name:  "Peter",
		Basic: []*Basic{&b, b1.(*Basic)},
	}
	instance = s.ParseResourceModel(&hm)

	j1 = instance.ToJsonapiModel()
	jsonapi.MarshalPayloadWithoutIncluded(os.Stdout, j1)

	// test hasSelf relations
	s, err = r.RegisterResource(reflect.TypeOf(HasSelf{}))
	require.Nil(t, err)

	hs := HasSelf{
		Id:   1,
		Name: "Peter",
		Self: nil,
	}
	instance = s.ParseResourceModel(&hs)

	j1 = instance.ToJsonapiModel()
	jsonapi.MarshalPayloadWithoutIncluded(os.Stdout, j1)

	// test circular relations
	s, err = r.RegisterResource(reflect.TypeOf(CircularHas{}))
	require.Nil(t, err)

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
	instance = s.ParseResourceModel(&ha)

	j1 = instance.ToJsonapiModel()
	jsonapi.MarshalPayloadWithoutIncluded(os.Stdout, j1)

	p1 := instance.ToResourceModel()
	bytes, err := json.Marshal(p1)
	require.Nil(t, err)
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
	instance = s.ParseResourceModel(&ha)

	j1 = instance.ToJsonapiModel()
	jsonapi.MarshalPayloadWithoutIncluded(os.Stdout, j1)

	p1 = instance.ToPGModel()
	bytes, err = json.Marshal(p1)
	require.Nil(t, err)
	println(string(bytes))
}
