// +build integration

package integration

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type defaultAttribute struct {
	Id   int64
	Name *string `jargo:",default:'John Doe'"`
}

// TestDefaultValues tests the behaviour of attributes with a default value.
func TestDefaultValues(t *testing.T) {
	resource, err := app.RegisterResource(defaultAttribute{})
	require.Nil(t, err)

	res, err := resource.InsertInstance(app.DB(), &defaultAttribute{}).Result()
	require.Nil(t, err)
	inserted := res.(*defaultAttribute)

	require.Equal(t, "John Doe", *inserted.Name)
}

type notnullWithoutDefault struct {
	Id   int64
	Name *string `jargo:",notnull"`
}

// TestNotnullWithoutDefault asserts that the notnull option is only
// allowed on fields with a default value.
func TestNotnullWithoutDefault(t *testing.T) {
	_, err := app.RegisterResource(notnullWithoutDefault{})
	require.EqualError(t, err, `"notnull" option may only be used in conjunction with the "default" option. use a primitive type instead`)
}
