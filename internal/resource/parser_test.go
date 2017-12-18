package resource

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

type Empty struct{}

type InvalidIdType struct {
	Id string
}

type UnannotatedId struct {
	Id int64
}

type IdOnly struct {
	Id int64 `jargo:""`
}

type InvalidTableName struct {
	Id int64 `jargo:",table:äöü"`
}

type InvalidTypeName0 struct {
	Id int64 `jargo:"-asdf"`
}

type InvalidTypeName1 struct {
	Id int64 `jargo:"asd$$f"`
}

type ValidTypeName struct {
	Id int64 `jargo:"resources,table:tbl_resources"`
}

type DisallowedColumnOptionId struct {
	Id int64 `jargo:"resources,column:asdf"`
}

type Simple0 struct {
	Id       int64 `jargo:""`
	Property string
}

type Simple1 struct {
	Id       int64  `jargo:""`
	Property string `jargo:"prop"`
}

type Simple2 struct {
	Id       int64  `jargo:""`
	Property string `jargo:"prop,column:col_property"`
}

type UnexportedProperty struct {
	Id       int64  `jargo:""`
	Property string `jargo:"-"`
}

type InvalidProperty0 struct {
	Id    int64 `jargo:""`
	Child BelongsTo0
}

type InvalidProperty1 struct {
	Id       int64 `jargo:""`
	Children []BelongsTo0
}

type InvalidProperty2 struct {
	Id       int64 `jargo:""`
	Children *[]BelongsTo0
}

type InvalidBelongsToType struct {
	Id      int64      `jargo:""`
	Parents []*HasOne0 `jargo:",belongsTo"`
}

type InvalidMany2ManyType struct {
	Id     int64  `jargo:""`
	Genres *Genre `jargo:",many2many:some_table"`
}

type Many2ManyMissingJoinTable struct {
	Id     int64    `jargo:""`
	Genres []*Genre `jargo:",many2many"`
}

type MissingRelationTag struct {
	Id    int64 `jargo:""`
	Child *BelongsTo0
}

type DisallowedColumnOptionRelation struct {
	Id    int64       `jargo:""`
	Child *BelongsTo0 `jargo:",has,column:asdf"`
}

// valid relationship models
type HasOne0 struct {
	Id    int64       `jargo:""`
	Child *BelongsTo0 `jargo:",has:Parent"`
}

type BelongsTo0 struct {
	Id     int64    `jargo:""`
	Parent *HasOne0 `jargo:",belongsTo"`
}

// valid many2many relationship models
type Book struct {
	Id     int64    `jargo:""`
	Genres []*Genre `jargo:",many2many:book_genres"`
}

type Genre struct {
	Id    int64   `jargo:""`
	Books []*Book `jargo:",many2many:book_genres"`
}

// other options
type Options0 struct {
	Id       int64  `jargo:""`
	Property string `jargo:",notnull,readonly,sort:false"`
}

type InvalidBoolOption struct {
	Id       int64  `jargo:""`
	Property string `jargo:",sort:asdf"`
}

type InvalidOption struct {
	Id       int64  `jargo:""`
	Property string `jargo:",someUnknownOption:asdf"`
}

func TestParseResourceStruct(t *testing.T) {
	_, err := parseResourceStruct("hi")
	assert.EqualError(t, err, errInvalidModelType.Error())

	_, err = parseResourceStruct(Empty{})
	assert.EqualError(t, err, errMissingIdField.Error())

	_, err = parseResourceStruct(InvalidIdType{})
	assert.EqualError(t, err, errInvalidIdType.Error())

	_, err = parseResourceStruct(UnannotatedId{})
	assert.EqualError(t, err, errUnannotatedIdField.Error())

	// test default resource name generation
	rd, err := parseResourceStruct(IdOnly{})
	assert.Nil(t, err)
	assert.Equal(t, "id_onlys", rd.name, "resource name is not camel-cased, pluralized version of struct name")
	assert.Equal(t, "id_onlys", rd.table, "table name is not camel-cased, pluralized version of struct name")

	_, err = parseResourceStruct(InvalidTableName{})
	assert.EqualError(t, err, errInvalidTableName.Error())

	_, err = parseResourceStruct(InvalidTypeName0{})
	assert.EqualError(t, err, errInvalidMemberName.Error())

	_, err = parseResourceStruct(InvalidTypeName1{})
	assert.EqualError(t, err, errInvalidMemberName.Error())

	rd, err = parseResourceStruct(ValidTypeName{})
	assert.Nil(t, err)
	assert.Equal(t, "resources", rd.name)
	assert.Equal(t, "tbl_resources", rd.table)

	_, err = parseResourceStruct(DisallowedColumnOptionId{})
	assert.EqualError(t, err, errDisallowedOption(optionColumn).Error())

	_, err = parseResourceStruct(DisallowedColumnOptionRelation{})
	assert.EqualError(t, err, errDisallowedOption(optionColumn).Error())

	// test simple attribute field parsing
	rd, err = parseResourceStruct(Simple0{})
	assert.Nil(t, err)
	assertId(t, rd)
	assertAttribute(t, rd, "Property", "property", "property")

	rd, err = parseResourceStruct(Simple1{})
	assert.Nil(t, err)
	assertAttribute(t, rd, "Property", "prop", "prop")

	rd, err = parseResourceStruct(Simple2{})
	assert.Nil(t, err)
	assertAttribute(t, rd, "Property", "prop", "col_property")

	// test parsing of unexported properties
	rd, err = parseResourceStruct(UnexportedProperty{})
	assert.Nil(t, err)
	assertAttribute(t, rd, "Property", "-", "property")

	// parse invalid field types
	_, err = parseResourceStruct(InvalidProperty0{})
	assert.EqualError(t, err, errStructType.Error())

	_, err = parseResourceStruct(InvalidProperty1{})
	assert.EqualError(t, err, errMissingRelationTag.Error())

	_, err = parseResourceStruct(InvalidProperty2{})
	assert.EqualError(t, err, errMissingRelationTag.Error())

	// test relation parsing
	_, err = parseResourceStruct(MissingRelationTag{})
	assert.EqualError(t, err, errMissingRelationTag.Error())

	_, err = parseResourceStruct(Many2ManyMissingJoinTable{})
	assert.EqualError(t, err, errMissingMany2ManyJoinTable.Error())

	_, err = parseResourceStruct(InvalidBelongsToType{})
	assert.EqualError(t, err, errInvalidBelongsToType.Error())

	_, err = parseResourceStruct(InvalidMany2ManyType{})
	assert.EqualError(t, err, errInvalidMany2ManyType.Error())

	rd, err = parseResourceStruct(HasOne0{})
	assert.Nil(t, err)
	assertHas(t, rd, "Child", "child", "Parent")

	rd, err = parseResourceStruct(BelongsTo0{})
	assert.Nil(t, err)
	assertBelongsTo(t, rd, "Parent", "parent")

	rd, err = parseResourceStruct(Book{})
	assert.Nil(t, err)
	assertMany2Many(t, rd, "Genres", "genres", "book_genres")

	rd, err = parseResourceStruct(Genre{})
	assert.Nil(t, err)
	assertMany2Many(t, rd, "Books", "books", "book_genres")

	// test readonly, sort, filter, notnull, unique
	rd, err = parseResourceStruct(Options0{})
	assert.Nil(t, err)
	assertField(t, rd, "Property", attribute, "property", "property",
		true, false, true, true, false, "", "", "")

	_, err = parseResourceStruct(InvalidBoolOption{})
	assert.Error(t, err) // strconv.parseBool error

	_, err = parseResourceStruct(InvalidOption{})
	assert.EqualError(t, err, errDisallowedOption("someUnknownOption").Error())
}

func assertField(t *testing.T, rd *resourceDefinition, goName string, typ fieldType,
	name string, column string, readonly bool, sort bool, filter bool,
	sqlNotnull bool, sqlUnique bool, sqlDefault string,
	pgFk string, pgJoinTable string) {

	assert.NotNil(t, rd)

	for _, field := range rd.fields {
		if field.structField.Name != goName {
			continue
		}

		assert.Equal(t, typ, field.typ)
		assert.Equal(t, name, field.name)
		assert.Equal(t, column, field.column)
		assert.Equal(t, readonly, field.readonly)
		assert.Equal(t, sort, field.sort)
		assert.Equal(t, filter, field.filter)
		assert.Equal(t, sqlNotnull, field.sqlNotnull)
		assert.Equal(t, sqlUnique, field.sqlUnique)
		assert.Equal(t, sqlDefault, field.sqlDefault)
		assert.Equal(t, pgFk, field.pgFk)
		assert.Equal(t, pgJoinTable, field.pgJoinTable)
		return
	}

	t.Errorf("could not find field with name %s", goName)
}

func assertId(t *testing.T, rd *resourceDefinition) {
	assertField(t, rd, "Id", id, "", "", false, false, false, false, false, "", "", "")
}

func assertAttribute(t *testing.T, rd *resourceDefinition, goName string, name string, column string) {
	assertField(t, rd, goName, attribute, name, column, false, true, true, false, false, "", "", "")
}

func assertHas(t *testing.T, rd *resourceDefinition, goName string, name string, fk string) {
	assertField(t, rd, goName, has, name, "", false, true, true, false, false, "", fk, "")
}

func assertBelongsTo(t *testing.T, rd *resourceDefinition, goName string, name string) {
	assertField(t, rd, goName, belongsTo, name, "", false, true, true, false, false, "", "", "")
}

func assertMany2Many(t *testing.T, rd *resourceDefinition, goName string, name string, joinTable string) {
	assertField(t, rd, goName, many2many, name, "", false, true, true, false, false, "", "", joinTable)
}
