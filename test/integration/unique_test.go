// +build integration

package integration

import (
	"github.com/crushedpixel/jargo"
	"github.com/stretchr/testify/require"
	"testing"
)

type uniqueAttribute struct {
	Id       int64
	UserName string `jargo:",unique"`
}

// TestUniqueAttributes tests the behaviour of unique attribute fields.
func TestUniqueAttributes(t *testing.T) {
	resource, err := app.RegisterResource(uniqueAttribute{})
	require.Nil(t, err)

	original := &uniqueAttribute{UserName: "Marius"}

	// insert instance
	_, err = resource.InsertInstance(app.DB(), original).Result()
	require.Nil(t, err)

	// insert instance again
	_, err = resource.InsertInstance(app.DB(), original).Result()
	// query must return UniqueViolationError
	require.IsType(t, &jargo.UniqueViolationError{}, err)

	uve := err.(*jargo.UniqueViolationError)
	require.Equal(t, "user_name", uve.Column)
	require.Equal(t, "user-name", uve.Field)
}
