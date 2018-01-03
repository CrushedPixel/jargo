// +build integration

package integration

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Model struct {
	Id int64 `jargo:"models"`

	CreatedAt time.Time `jargo:",createdAt"`
	UpdatedAt time.Time `jargo:",updatedAt"`

	Name string
}

func TestAutoTimestamps(t *testing.T) {
	modelResource, err := app.RegisterResource(Model{})
	require.Nil(t, err)

	r, err := modelResource.InsertOne(app.DB, &Model{Name: "A"}).Result()
	require.Nil(t, err)
	instance := r.(*Model)
	// instance.CreatedAt and instance.UpdatedAt should have been populated by the server
	assert.NotEmpty(t, instance.CreatedAt)
	assert.NotEmpty(t, instance.UpdatedAt)
	updatedBefore := instance.UpdatedAt

	// wait a short amount of time to ensure
	// the timestamp of the update is going to be different
	time.Sleep(time.Millisecond * 10)

	instance.Name = "B"
	r, err = modelResource.UpdateOne(app.DB, instance).Result()
	require.Nil(t, err)
	instance = r.(*Model)
	// instance.UpdatedAt should be timestamp of the update as set by the server
	assert.NotEmpty(t, instance.CreatedAt)
	assert.NotEmpty(t, instance.UpdatedAt)
	assert.NotEqual(t, updatedBefore, instance.UpdatedAt)
}
