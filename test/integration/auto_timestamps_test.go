// +build integration

package integration

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type AutoTimestamps struct {
	Id int64 `jargo:"auto_timestamps,alias:auto_timestamp"`

	CreatedAt time.Time `jargo:",createdAt"`
	UpdatedAt time.Time `jargo:",updatedAt"`

	Name string
}

func TestAutoTimestamps(t *testing.T) {
	resource, err := app.RegisterResource(AutoTimestamps{})
	require.Nil(t, err)

	// note: although the db.OnQueryProcessed logger prints a query
	// indicating the createdAt timestamp is set on the client,
	// it actually executes a different query, containing DEFAULT values
	// for createdAt and updatedAt.
	r, err := resource.InsertOne(app.DB, &AutoTimestamps{Name: "A"}).Result()
	require.Nil(t, err)
	instance := r.(*AutoTimestamps)
	assert.Equal(t, "A", instance.Name)
	// instance.CreatedAt and instance.UpdatedAt should have been populated by the server
	assert.NotEmpty(t, instance.CreatedAt)
	assert.NotEmpty(t, instance.UpdatedAt)
	assert.Equal(t, instance.CreatedAt, instance.UpdatedAt)

	// wait a short amount of time to ensure
	// the timestamp of the update is going to be different
	time.Sleep(time.Millisecond * 10)

	// update resource
	instance.Name = "B"
	r, err = resource.UpdateOne(app.DB, instance).Result()
	require.Nil(t, err)
	updated := r.(*AutoTimestamps)
	assert.Equal(t, "B", updated.Name)

	// instance.CreatedAt should not have changed
	assert.NotEmpty(t, updated.CreatedAt)
	assert.Equal(t, instance.CreatedAt, updated.CreatedAt)

	// instance.UpdatedAt should be updated timestamp
	assert.NotEmpty(t, updated.UpdatedAt)
	assert.NotEqual(t, instance.UpdatedAt, updated.UpdatedAt)
}
