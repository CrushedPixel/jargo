// +build integration

package integration

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type nullableRelation struct {
	Id       int64
	Relation *dummy `jargo:",belongsTo,nullable"`
}

type nonNullableRelation struct {
	Id       int64
	Relation *dummy `jargo:",belongsTo"`
}

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
		&nonNullableRelation{Relation: dummyInstance}).
		Result()
	require.Nil(t, err)
	require.Equal(t, res.(*nonNullableRelation).Relation.Id, dummyInstance.Id)

	// insert nonNullableRelation instance with relation set to null
	_, err = resource.InsertInstance(app.DB(), &nonNullableRelation{}).Result()
	require.EqualError(t, err, "encountered null value on belongsTo relation not marked nullable")
}
