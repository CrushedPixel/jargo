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

type nullableRelation struct {
	Id       int64
	Relation *dummy `jargo:",belongsTo"`
}

type nonNullableRelation struct {
	Id       int64
	Relation dummy `jargo:",belongsTo"`
}

type nullableIdField struct {
	Id *int64 // this is invalid
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
	require.NotEqual(t, reflect.ValueOf(inserted.Age).Pointer(), reflect.ValueOf(original.Age).Pointer())

	// fetch created resource from database to ensure data was properly stored
	res, err = resource.SelectById(app.DB(), inserted.Id).Result()
	require.Nil(t, err)
	fetched := res.(*nullableAttribute)

	require.Equal(t, fetched.Age, original.Age)
	require.Equal(t, fetched.Name, original.Name)
}

/* TODO: re-introduce this check when implementing support for non-int64 id fields
// TestNullableIdFields tests the behaviour of nullable id fields (which is invalid).
func TestNullableIdFields(t *testing.T) {
	_, err := app.RegisterResource(nullableIdField{})
	require.EqualError(t, err, "id field must not be nullable")
}
*/

// TestNullableRelations tests the behaviour of nullable relation fields.
func TestNullableRelations(t *testing.T) {
	resource, err := app.RegisterResource(nullableRelation{})
	require.Nil(t, err)

	// insert nullableRelation instance with relation set to null
	_, err = resource.InsertInstance(app.DB(), &nullableRelation{}).Result()
	require.Nil(t, err)

	// insert nullableRelation instance with relation set to value
	res, err := resource.InsertInstance(app.DB(),
		&nullableRelation{Relation: dummyInstance}).
		Result()
	require.Nil(t, err)
	require.Equal(t, res.(*nullableRelation).Relation.Id, dummyInstance.Id)
}

// TestNonNullableRelations tests the behaviour of non-nullable (default)
// relation fields.
func TestNonNullableRelations(t *testing.T) {
	resource, err := app.RegisterResource(nonNullableRelation{})
	require.Nil(t, err)

	// insert nonNullableRelation instance with relation set to value
	res, err := resource.InsertInstance(app.DB(),
		&nonNullableRelation{Relation: *dummyInstance}).
		Result()
	require.Nil(t, err)
	require.Equal(t, res.(*nonNullableRelation).Relation.Id, dummyInstance.Id)

	// insert nonNullableRelation instance with relation set to null
	_, err = resource.InsertInstance(app.DB(), &nonNullableRelation{}).Result()
	require.EqualError(t, err, "encountered null value on belongsTo relation not marked nullable")
}
