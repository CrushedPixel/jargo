// +build integration

package integration

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type nonNullableSlice struct {
	Id    int64 `jargo:"non-nullable-slices,table:non_nullable_slices,alias:non_nullable_slice"`
	Slice []string
}

// TestSlices tests the behaviour of slice attributes.
func TestSlices(t *testing.T) {
	resource, err := app.RegisterResource(nonNullableSlice{})
	require.Nil(t, err)

	// insert nonNullableSlice instance with slice set to nil
	res, err := resource.InsertInstance(app.DB(),
		&nonNullableSlice{}).Result()
	// expect nil slice being treated like an empty slice
	require.Nil(t, err)

	// select slice from database
	res, err = resource.SelectById(app.DB(),
		res.(*nonNullableSlice).Id).Result()
	require.Nil(t, err)
	var nilSlice []string
	require.Equal(t, nilSlice, res.(*nonNullableSlice).Slice)

	// insert slice with values
	valueSlice := []string{"hello", "world"}
	res, err = resource.InsertInstance(app.DB(),
		&nonNullableSlice{Slice: valueSlice}).Result()
	require.Nil(t, err)
	val := res.(*nonNullableSlice)

	require.Equal(t, "hello", val.Slice[0])
	require.Equal(t, "world", val.Slice[1])
}
