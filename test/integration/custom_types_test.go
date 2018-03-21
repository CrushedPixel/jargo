// +build integration

package integration

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type ageType int8

type myStruct struct {
	Key   string
	Value string
}

type mySlice []string

type customTypeAttribute struct {
	Id                 int64
	Age                ageType
	NullableAge        *ageType
	AnotherNullableAge *ageType
	Struct             myStruct
	Slice              mySlice
}

// TestCustomTypes tests the behaviour of attributes with custom types.
func TestCustomTypes(t *testing.T) {
	resource, err := app.RegisterResource(customTypeAttribute{})
	require.Nil(t, err)

	a0 := ageType(10)
	a1 := ageType(15)

	original := &customTypeAttribute{
		Age:         a0,
		NullableAge: &a1,
		Struct: myStruct{
			Key:   "Hello",
			Value: "World",
		},
		Slice: mySlice{"Lorem", "Ipsum"},
	}
	res, err := resource.InsertInstance(app.DB(), original).Result()
	require.Nil(t, err)
	inserted := res.(*customTypeAttribute)

	// ensure instance properly encodes to json
	json, err := resource.ResponseAllFields(inserted).Payload()
	require.Nil(t, err)
	require.Equal(t,
		`{"data":{"type":"custom_type_attributes","id":"1","attributes":{"age":10,"another-nullable-age":null,"nullable-age":15,"slice":["Lorem","Ipsum"],"struct":{"Key":"Hello","Value":"World"}}}}`,
		json)
}
