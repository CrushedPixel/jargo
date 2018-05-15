// +build integration

package integration

import (
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

type unmarshalTest struct {
	Id   uuid.UUID
	Name string
	Age  int
}

func TestPayloadUnmarshaling(t *testing.T) {
	resource, err := app.RegisterResource(unmarshalTest{})
	require.Nil(t, err)

	res, err := resource.ParseJsonapiPayloadString(`{"data":{"type":"unmarshal-tests","id":"159b6f97-ed4b-4b52-b9cb-323754f948ad","attributes":{"name":"Steve","age":10}}}`, nil, false)
	require.Nil(t, err)
	val := res.(*unmarshalTest)

	require.Equal(t, uuid.FromStringOrNil("159b6f97-ed4b-4b52-b9cb-323754f948ad"), val.Id)
	require.Equal(t, "Steve", val.Name)
	require.Equal(t, 10, val.Age)
}
