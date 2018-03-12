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

type defaultOnNonPointer struct {
	Id   int64
	Name string `jargo:",default:'John Doe'"`
}

// TestDefaultOnNonPointer asserts that the default option is not
// allowed on non-pointer types.
func TestDefaultOnNonPointer(t *testing.T) {
	_, err := app.RegisterResource(defaultOnNonPointer{})
	require.EqualError(t, err, `"default" option may only be used on pointer types`)
}
