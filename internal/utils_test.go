package internal

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestPluralize(t *testing.T) {
	val := pluralize("")
	assert.Equal(t, "", val)

	val = pluralize("resource")
	assert.Equal(t, "resources", val)

	val = pluralize("resources")
	assert.Equal(t, "resources", val)
}

func TestIsValidJsonapiMemberName(t *testing.T) {
	ok := isValidJsonapiMemberName("")
	assert.Equal(t, false, ok)

	ok = isValidJsonapiMemberName("h")
	assert.Equal(t, true, ok)

	ok = isValidJsonapiMemberName("-hi")
	assert.Equal(t, false, ok)

	ok = isValidJsonapiMemberName("hi-")
	assert.Equal(t, false, ok)

	ok = isValidJsonapiMemberName("$hi")
	assert.Equal(t, false, ok)

	ok = isValidJsonapiMemberName("hi$")
	assert.Equal(t, false, ok)

	ok = isValidJsonapiMemberName("hi$hi")
	assert.Equal(t, false, ok)

	ok = isValidJsonapiMemberName("hi hi")
	assert.Equal(t, false, ok)

	ok = isValidJsonapiMemberName("test_01")
	assert.Equal(t, true, ok)

	ok = isValidJsonapiMemberName("Test01")
	assert.Equal(t, true, ok)
}

func TestIsValidSQLName(t *testing.T) {
	ok := isValidJsonapiMemberName("")
	assert.Equal(t, false, ok)

	ok = isValidSQLName("%hi")
	assert.Equal(t, false, ok)

	ok = isValidSQLName("hi")
	assert.Equal(t, true, ok)

	ok = isValidSQLName("_hi")
	assert.Equal(t, true, ok)
}
