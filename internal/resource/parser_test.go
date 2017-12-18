package resource

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

type EmptyTestModel struct{}

type InvalidIdTypeTestModel struct {
	Id string
}

type UnannotatedIdTestModel struct {
	Id int64
}

type IdOnlyTestModel struct {
	Id int64 `jargo:""`
}

type InvalidTableNameTestModel struct {
	Id int64 `jargo:",table:äöü"`
}

type InvalidTypeNameTestModel0 struct {
	Id int64 `jargo:"-asdf"`
}

type InvalidTypeNameTestModel1 struct {
	Id int64 `jargo:"asd$$f"`
}

type ValidTypeNameTestModel struct {
	Id int64 `jargo:"resources,table:tbl_resources"`
}

type SimpleTestModel0 struct {
	Id       int64 `jargo:""`
	Property string
}

type SimpleTestModel1 struct {
	Id       int64  `jargo:""`
	Property string `jargo:"prop"`
}

type SimpleTestModel2 struct {
	Id       int64  `jargo:""`
	Property string `jargo:"prop,column:col_property"`
}

func TestParseResourceStruct(t *testing.T) {
	_, err := parseResourceStruct("hi")
	assert.Error(t, errInvalidModelType, err)

	_, err = parseResourceStruct(EmptyTestModel{})
	assert.Error(t, errMissingIdField, err)

	_, err = parseResourceStruct(InvalidIdTypeTestModel{})
	assert.Error(t, errInvalidIdType, err)

	_, err = parseResourceStruct(UnannotatedIdTestModel{})
	assert.Error(t, errUnannotatedIdField, err)

	// test default resource name generation
	rd, err := parseResourceStruct(IdOnlyTestModel{})
	assert.Equal(t, "id_only_test_models", rd.name, "resource name is not camel-cased, pluralized version of struct name")
	assert.Equal(t, "id_only_test_models", rd.table, "table name is not camel-cased, pluralized version of struct name")

	_, err = parseResourceStruct(InvalidTableNameTestModel{})
	assert.Error(t, errInvalidTableName, err)

	_, err = parseResourceStruct(InvalidTypeNameTestModel0{})
	assert.Error(t, errInvalidMemberName, err)

	_, err = parseResourceStruct(InvalidTypeNameTestModel1{})
	assert.Error(t, errInvalidMemberName, err)

	rd, err = parseResourceStruct(ValidTypeNameTestModel{})
	assert.Equal(t, "resources", rd.name)
	assert.Equal(t, "tbl_resources", rd.table)

	// test simple attribute field parsing
	rd, err = parseResourceStruct(SimpleTestModel0{})
	assert.Contains(t, rd.fields, &field{
		name:   "property",
		column: "property",
	})

	rd, err = parseResourceStruct(SimpleTestModel2{})
	assert.Contains(t, rd.fields, &field{
		name:   "prop",
		column: "col_property",
	})

	// TODO: tests for rest of field properties
}
