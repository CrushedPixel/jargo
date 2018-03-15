// +build integration

package integration

import (
	"fmt"
	"github.com/crushedpixel/jargo"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

type uuidIdField struct {
	Id uuid.UUID
}

// TestUUID_Id tests the behaviour of UUID id fields.
func TestUUID_Id(t *testing.T) {
	resource, err := app.RegisterResource(uuidIdField{})
	require.Nil(t, err)

	// insert multiple instances of customIdTypeB
	// to ensure auto-generated ids work for uuid id fields
	for i := 0; i < 5; i++ {
		res, err := resource.InsertInstance(app.DB(), &uuidIdField{}).Result()
		require.Nil(t, err)

		inserted := res.(*uuidIdField)
		require.NotZero(t, inserted.Id)

		// ensure instance properly encodes to json
		json, err := resource.ResponseAllFields(inserted).Payload()
		require.Nil(t, err)
		require.Equal(t,
			fmt.Sprintf(`{"data":{"type":"uuid_id_field","id":"%s"}}`, inserted.Id),
			json)
	}
}

type nullableIdField struct {
	Id *int64
}

// TestNullableIdFields ensures that pointer types are not allowed for id fields.
func TestNullableIdFields(t *testing.T) {
	_, err := app.RegisterResource(nullableIdField{})
	require.EqualError(t, err, "pointer types are not allowed for id fields")
}

type myString string
type myInt int

type customIdTypeA struct {
	Id myString
}

type customIdTypeB struct {
	Id   myInt
	Attr string
}

// TestCustomIdTypes tests the behaviour of id fields with custom types for the id field.
func TestCustomIdTypes(t *testing.T) {
	resourceA, err := app.RegisterResource(customIdTypeA{})
	require.Nil(t, err)

	resourceB, err := app.RegisterResource(customIdTypeB{})
	require.Nil(t, err)

	originalA := &customIdTypeA{Id: "Marius"}
	// insert instance of customIdTypeA
	res, err := resourceA.InsertInstance(app.DB(), originalA).Result()
	require.Nil(t, err)
	inserted := res.(*customIdTypeA)
	// ensure id was properly set
	require.Equal(t, originalA.Id, inserted.Id)

	// ensure instance properly encodes to json
	json, err := resourceA.ResponseAllFields(inserted).Payload()
	require.Nil(t, err)
	require.Equal(t,
		`{"data":{"type":"custom_id_type_as","id":"Marius"}}`,
		json)

	// insert same instance again, expecting pk constraint violation
	_, err = resourceA.InsertInstance(app.DB(), originalA).Result()
	// query must return UniqueViolationError
	require.IsType(t, &jargo.UniqueViolationError{}, err)

	uve := err.(*jargo.UniqueViolationError)
	require.Equal(t, "id", uve.Column)
	require.Equal(t, "id", uve.Field)

	// insert multiple instances of customIdTypeB
	// to ensure auto-incrementing ids work for custom number types
	for i := 0; i < 5; i++ {
		res, err := resourceB.InsertInstance(app.DB(), &customIdTypeB{Attr: "hi"}).Result()
		require.Nil(t, err)

		inserted := res.(*customIdTypeB)
		require.Equal(t, i+1, inserted.Id)
		require.Equal(t, "hi", inserted.Attr)

		// ensure instance properly encodes to json
		json, err := resourceB.ResponseAllFields(inserted).Payload()
		require.Nil(t, err)
		require.Equal(t,
			fmt.Sprintf(`{"data":{"type":"custom_id_type_bs","id":"%d","attributes":{"attr":"hi"}}}`, i+1),
			json)
	}
}
