package parser

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseJargoTag(t *testing.T) {
	p := ParseJargoTag("")
	assert.Equal(t, p.Name, "")

	p = ParseJargoTag("myName,key0,key1:value1,key2:value2,key3:")
	assert.NotNil(t, p, "Parsed tag is nil")

	assert.Equal(t, p.Name, "myName", "Parsed name is incorrect")
	assertKeyValue(t, p.Options, "key0", "")
	assertKeyValue(t, p.Options, "key1", "value1")
	assertKeyValue(t, p.Options, "key2", "value2")
	assertKeyValue(t, p.Options, "key3", "")
	assertKeyValue(t, p.Options, "key4", "")
}

func TestParseJargoTagDefaultName(t *testing.T) {
	p := ParseJargoTagDefaultName("", "default")
	assert.Equal(t, p.Name, "default")

	p = ParseJargoTagDefaultName("myName", "default")
	assert.Equal(t, p.Name, "myName")
}

func assertKeyValue(t *testing.T, options map[string]string, key string, value string) {
	assert.Equalf(t, options[key], value, "Parsed option value for %s is incorrect", key)
}
