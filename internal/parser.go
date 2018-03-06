package internal

import (
	"errors"
	"fmt"
	"github.com/c9s/inflect"
	"reflect"
)

const (
	idFieldName         = "Id"
	unexportedFieldName = "-"

	jargoFieldTag   = "jargo"
	optionTable     = "table"
	optionAlias     = "alias"
	optionColumn    = "column"
	optionHas       = "has"
	optionBelongsTo = "belongsTo"
	optionMany2Many = "many2many"
	optionReadonly  = "readonly"
	optionSort      = "sort"
	optionFilter    = "filter"
	optionOmitempty = "omitempty"
	optionNotnull   = "notnull"
	optionUnique    = "unique"
	optionDefault   = "default"
	optionCreatedAt = "createdAt"
	optionUpdatedAt = "updatedAt"
)

type fieldType int

const (
	attribute fieldType = iota
	has
	belongsTo
	many2many
)

var (
	idFieldType = reflect.TypeOf(int64(0))

	errInvalidModelType          = errors.New("model has to be a struct")
	errMissingIdField            = errors.New("missing id field")
	errInvalidIdType             = errors.New("id field must be of type int")
	errInvalidMemberName         = errors.New(`member name has to adhere to the jsonapi specification and not include characters marked as "not recommended"`)
	errInvalidTableName          = errors.New("table name may only consist of [0-9,a-z,A-Z$_]")
	errInvalidTableAlias         = errors.New("alias may only consist of [0-9,a-z,A-Z$_]")
	errAliasEqualsTableName      = errors.New("alias may not be equal to table name")
	errMissingMany2ManyJoinTable = errors.New("missing many2many join table definition")
	errMultipleRelationTypes     = errors.New("multiple relation type options are not allowed")

	errDisallowedOption = func(option string) error {
		return errors.New(fmt.Sprintf(`option "%s" is not allowed here`, option))
	}
)

// parses a schema definition from a resource model type.
func (r SchemaRegistry) newSchemaDefinition(t reflect.Type) *Schema {
	if t.Kind() != reflect.Struct {
		panic(errInvalidModelType)
	}

	schema := parseSchema(t)

	// parse struct fields
	var jsonapiJoinFields []reflect.StructField
	var pgJoinFields []reflect.StructField
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		field := r.parseField(schema, &f)

		schema.fields = append(schema.fields, field)
		jsonapiJoinFields = append(jsonapiJoinFields, field.jsonapiJoinFields()...)
		pgJoinFields = append(pgJoinFields, field.pgJoinFields()...)
	}

	schema.joinJsonapiModelType = reflect.StructOf(jsonapiJoinFields)
	schema.joinPGModelType = reflect.StructOf(pgJoinFields)
	return schema
}

func (r SchemaRegistry) generateSchemaModels(schema *Schema) {
	var jsonapiFields []reflect.StructField
	var pgFields []reflect.StructField
	for _, f := range schema.fields {
		jsonapiFields = append(jsonapiFields, f.jsonapiFields()...)
		pgFields = append(pgFields, f.pgFields()...)
	}

	schema.jsonapiModelType = reflect.StructOf(jsonapiFields)
	schema.pgModelType = reflect.StructOf(pgFields)
}

// parses a struct's id field, retrieving
// general schema information from the jargo struct tag.
func parseSchema(t reflect.Type) *Schema {
	f, ok := t.FieldByName(idFieldName)
	if !ok {
		panic(errMissingIdField)
	}
	if f.Type != idFieldType {
		panic(errInvalidIdType)
	}

	// parse jargo struct tag
	tag := f.Tag.Get(jargoFieldTag)

	// parse schema name, sql table and sql alias
	// from struct tag.
	// default sql alias is the snake_cased struct name.
	// default name and sql table is the pluralized version
	// of the alias.
	// "UserProfile" => "user_profile", "user_profiles"
	singleName := inflect.Underscore(t.Name())
	defaultName := inflect.Pluralize(singleName)
	parsed := parseJargoTagDefaultName(tag, defaultName)

	schema := &Schema{
		name:              parsed.Name,
		table:             parsed.Name,
		alias:             singleName,
		resourceModelType: t,
	}

	// parse options defined in struct tag.
	// they may be used to override sql table and alias.
	for option, value := range parsed.Options {
		switch option {
		case optionTable:
			schema.table = value
		case optionAlias:
			schema.alias = value
		default:
			panic(errDisallowedOption(option))
		}
	}

	// validate member name
	if !isValidJsonapiMemberName(schema.name) {
		panic(errInvalidMemberName)
	}

	// validate table name
	if !isValidSQLName(schema.table) {
		panic(errInvalidTableName)
	}

	// validate alias
	if !isValidSQLName(schema.alias) {
		panic(errInvalidTableAlias)
	}

	// ensure alias is not the same as table name
	if schema.alias == schema.table {
		panic(errAliasEqualsTableName)
	}

	return schema
}

// parses a struct field into a schema field.
// returns nil, nil for non-attribute fields.
func (r SchemaRegistry) parseField(schema *Schema, f *reflect.StructField) SchemaField {
	if f.Name == idFieldName {
		return newIdField(schema)
	}

	// determine field type from jargo tag
	parsed := parseJargoTag(f.Tag.Get(jargoFieldTag))

	typ := attribute
	var val string
	for option, value := range parsed.Options {
		switch option {
		case optionHas:
			if typ != attribute {
				panic(errMultipleRelationTypes)
			}
			typ = has
			val = value
		case optionBelongsTo:
			if typ != attribute {
				panic(errMultipleRelationTypes)
			}
			typ = belongsTo
		case optionMany2Many:
			if typ != attribute {
				panic(errMultipleRelationTypes)
			}
			typ = many2many
			val = value
		}
	}

	switch typ {
	case has:
		return newHasField(r, schema, f, val)
	case belongsTo:
		return newBelongsToField(r, schema, f)
	case many2many:
		// TODO: implement many2many relations
		panic(errors.New("many2many relations are not yet implemented"))
	default:
		return newAttrField(schema, f)
	}
}
