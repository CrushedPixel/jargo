// +build integration

package integration

import (
	"fmt"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

type autoIncrementingIdField struct {
	Id int64
}

// TestAutoIncrementingIdField tests the behaviour
// of auto-incrementing id fields.
func TestAutoIncrementingIdField(t *testing.T) {
	resource, err := app.RegisterResource(autoIncrementingIdField{})
	require.Nil(t, err)

	// insert multiple instances of customIdTypeB
	// to ensure auto-incrementing ids work for custom number types
	for i := 0; i < 5; i++ {
		res, err := resource.InsertInstance(app.DB(), &autoIncrementingIdField{}).Result()
		require.Nil(t, err)

		inserted := res.(*autoIncrementingIdField)
		require.Equal(t, int64(i+1), inserted.Id)

		// ensure instance properly encodes to json
		json, err := resource.ResponseAllFields(inserted).Payload()
		require.Nil(t, err)
		require.Equal(t,
			fmt.Sprintf(`{"data":{"type":"auto_incrementing_id_fields","id":"%d"}}`, i+1),
			json)
	}
}

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
			fmt.Sprintf(`{"data":{"type":"uuid_id_fields","id":"%s"}}`, inserted.Id),
			json)
	}
}

type invalidIdFieldA struct {
	Id *int64
}

type myString string
type invalidIdFieldB struct {
	Id myString
}

type myInt int
type invalidIdFieldC struct {
	Id myInt
}

// TestInvalidIdFields ensures the behaviour of non-supported id field types.
func TestInvalidIdFields(t *testing.T) {
	_, err := app.RegisterResource(invalidIdFieldA{})
	require.EqualError(t, err, "id field type must be string or non-float numeric, or implement encoding.TextMarshaler and encoding.TextUnmarshaler")

	_, err = app.RegisterResource(invalidIdFieldB{})
	require.EqualError(t, err, "id field type must be string or non-float numeric, or implement encoding.TextMarshaler and encoding.TextUnmarshaler")

	_, err = app.RegisterResource(invalidIdFieldC{})
	require.EqualError(t, err, "id field type must be string or non-float numeric, or implement encoding.TextMarshaler and encoding.TextUnmarshaler")
}
