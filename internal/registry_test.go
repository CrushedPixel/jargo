package internal

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"reflect"
)

// invalid inversions
/*
type InvalidInversion0 struct {
	Id int64 `jargo:""`
}

type InvalidInversion1 struct {
	Id    int64              `jargo:""`
	Owner *InvalidInversion0 `jargo:",belongsTo"`
}
*/

// valid models
type Human struct {
	Id     int64  `jargo:""`
	Name   string
	Age    int
	Gender bool
	Dogs   []*Dog `jargo:",has:Owner"`
}

type Dog struct {
	Id    int64  `jargo:""`
	Name  string
	Color string
	Owner *Human `jargo:",belongsTo"`
}

func TestRegistryInvalidInversions(t *testing.T) {
	// TODO: ensure inversions are valid
}

func TestRegistry(t *testing.T) {
	registry := NewRegistry()

	human, err := registry.RegisterResource(Human{})
	assert.Nil(t, err)

	dog, err := registry.RegisterResource(Dog{})
	assert.Nil(t, err)

	err = registry.InitializeResources()
	assert.Nil(t, err)

	// validate jsonapi and pg models
	jsonapiModel := human.jsonapiModel(allFields(human))
	assertStructField(t, jsonapiModel, "Id", `jsonapi:"primary,humans"`)
	assertStructField(t, jsonapiModel, "Name", `jsonapi:"attr,name"`)
	assertStructField(t, jsonapiModel, "Age", `jsonapi:"attr,age"`)
	assertStructField(t, jsonapiModel, "Gender", `jsonapi:"attr,gender"`)
	assertStructField(t, jsonapiModel, "Dogs", `jsonapi:"relation,dogs"`)

	assertStructField(t, human.pgModel, "TableName", `sql:"\"humans\",alias:\"human\""`)
	assertStructField(t, human.pgModel, "Id", `sql:"id,pk"`)
	assertStructField(t, human.pgModel, "Name", `sql:"name"`)
	assertStructField(t, human.pgModel, "Age", `sql:"age"`)
	assertStructField(t, human.pgModel, "Gender", `sql:"gender"`)
	assertStructField(t, human.pgModel, "Dogs", `pg:",fk:Owner"`)

	assertStructField(t, dog.pgModel, "OwnerId", "")
	assertStructField(t, dog.pgModel, "Owner", "")

	// TODO: test many2many and hasMany relations
	// TODO: test notnull, unique, default options
}

func assertStructField(t *testing.T, typ reflect.Type, name string, tag string) {
	f, ok := typ.FieldByName(name)
	assert.True(t, ok)
	assert.Equal(t, f.Tag, reflect.StructTag(tag))
}
