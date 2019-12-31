// +build integration

package integration

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type testModel struct {
	Id int64
}

func TestInsertCollection(t *testing.T) {
	resource, err := app.RegisterResource(testModel{})
	require.Nil(t, err)

	var instances []*testModel
	for i := 0; i < 5; i++ {
		instances = append(instances, &testModel{})
	}

	res, err := resource.InsertCollection(app.DB(), instances).Result()
	require.Nil(t, err)

	inserted := res.([]*testModel)
	require.Equal(t, len(inserted), len(instances))

	for i := 0; i < 5; i++ {
		// ensure the auto-incremented ids worked
		require.Equal(t, inserted[i].Id, i)
	}
}
