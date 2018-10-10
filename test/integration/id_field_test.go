// +build integration

package integration

import (
	"encoding/json"
	"fmt"
	"github.com/json-iterator/go"
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
			fmt.Sprintf(`{"data":{"type":"auto-incrementing-id-fields","id":"%d"}}`, i+1),
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
			fmt.Sprintf(`{"data":{"type":"uuid-id-fields","id":"%s"}}`, inserted.Id),
			json)

		// ensure the instance can be fetched
		// by the string representation of the UUID id
		res, err = resource.SelectById(app.DB(), inserted.Id.String()).Result()
		require.Nil(t, err)
		require.Equal(t,
			fmt.Sprintf(`{"data":{"type":"uuid-id-fields","id":"%s"}}`, inserted.Id),
			json)
	}
}

type MarshalerType struct {
	A string
	B int
	C bool
}

func (m *MarshalerType) MarshalText() ([]byte, error) {
	return json.Marshal(*m)
}

func (m *MarshalerType) UnmarshalText(text []byte) error {
	// we can't use encoding/json here directly.
	// see https://github.com/golang/go/issues/28119
	it := jsoniter.ParseBytes(jsoniter.ConfigDefault, text)

	it.ReadObject()
	m.A = it.ReadString()
	it.ReadObject()
	m.B = it.ReadInt()
	it.ReadObject()
	m.C = it.ReadBool()
	return it.Error
}

type marshalerIdField struct {
	Id MarshalerType
}

func TestTextMarshalerIdFields(t *testing.T) {
	resource, err := app.RegisterResource(marshalerIdField{})
	require.Nil(t, err)

	_, err = resource.InsertInstance(app.DB(), &marshalerIdField{
		Id: MarshalerType{
			A: "hello world",
			B: 10,
			C: true,
		},
	}).Result()
	require.Nil(t, err)

	res, err := resource.Select(app.DB()).Result()
	require.Nil(t, err)

	values := res.([]*marshalerIdField)
	require.Equal(t, 1, len(values))

	require.Equal(t, "hello world", values[0].Id.A)
	require.Equal(t, 10, values[0].Id.B)
	require.Equal(t, true, values[0].Id.C)
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
