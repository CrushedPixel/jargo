// +build integration

package integration

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

type nullableAttribute struct {
	Id   int64
	Name *string
	Age  *int
}

// TestNullableAttributes tests the behaviour of nullable attribute fields.
func TestNullableAttributes(t *testing.T) {
	resource, err := app.RegisterResource(nullableAttribute{})
	require.Nil(t, err)

	// insert instance with attribute set to null
	age := 20
	original := &nullableAttribute{
		Name: nil,
		Age:  &age,
	}
	res, err := resource.InsertInstance(app.DB(), original).Result()
	require.Nil(t, err)

	inserted := res.(*nullableAttribute)
	// ensure age pointer of returned instance points to different address
	require.NotEqual(t, reflect.ValueOf(original.Age).Pointer(), reflect.ValueOf(inserted.Age).Pointer())

	// fetch created resource from database to ensure data was properly stored
	res, err = resource.SelectById(app.DB(), inserted.Id).Result()
	require.Nil(t, err)
	fetched := res.(*nullableAttribute)

	require.Equal(t, original.Age, fetched.Age)
	require.Equal(t, original.Name, fetched.Name)

	// ensure instance properly encodes to json
	json, err := resource.ResponseAllFields(fetched).Payload()
	require.Nil(t, err)
	require.Equal(t,
		`{"data":{"type":"nullable-attributes","id":"1","attributes":{"age":20,"name":null}}}`,
		json)
}

type nullableRelation struct {
	Id       int64
	Relation *dummy `jargo:",belongsTo"`
}

// TestNullableRelations tests the behaviour of nullable relation fields.
func TestNullableRelations(t *testing.T) {
	resource, err := app.RegisterResource(nullableRelation{})
	require.Nil(t, err)

	// insert nullableRelation instance with relation set to null
	res, err := resource.InsertInstance(app.DB(), &nullableRelation{}).Result()
	require.Nil(t, err)
	// relation of returned resource instance should be nil
	require.Nil(t, res.(*nullableRelation).Relation)

	// insert nullableRelation instance with relation set to value
	res, err = resource.InsertInstance(app.DB(),
		&nullableRelation{Relation: dummyInstance}).
		Result()
	require.Nil(t, err)
	require.Equal(t, dummyInstance.Id, res.(*nullableRelation).Relation.Id)
}

type nonNullableRelation struct {
	Id       int64
	Relation dummy `jargo:",belongsTo"`
}

// TestNonNullableRelations tests the behaviour of non-nullable relation fields.
func TestNonNullableRelations(t *testing.T) {
	resource, err := app.RegisterResource(nonNullableRelation{})
	require.Nil(t, err)

	// insert nonNullableRelation instance with relation set to value
	res, err := resource.InsertInstance(app.DB(),
		&nonNullableRelation{Relation: *dummyInstance}).
		Result()
	require.Nil(t, err)
	require.Equal(t, dummyInstance.Id, res.(*nonNullableRelation).Relation.Id)
}
