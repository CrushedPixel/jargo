// +build integration

package integration

import (
	"fmt"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

type uuidAttribute struct {
	Id           int64
	Uuid         uuid.UUID
	NullableUuid *uuid.UUID
}

func TestUUIDAttributes(t *testing.T) {
	resource, err := app.RegisterResource(uuidAttribute{})
	require.Nil(t, err)

	uid, err := uuid.NewV4()
	require.Nil(t, err)

	original := &uuidAttribute{
		Uuid: uid,
	}
	res, err := resource.InsertInstance(app.DB(), original).Result()
	require.Nil(t, err)
	inserted := res.(*uuidAttribute)

	// ensure instance properly encodes to json
	json, err := resource.ResponseAllFields(inserted).Payload()
	require.Nil(t, err)
	require.Equal(t,
		fmt.Sprintf(`{"data":{"type":"uuid-attributes","id":"1","attributes":{"nullable-uuid":null,"uuid":"%s"}}}`, uid.String()),
		json)
}
