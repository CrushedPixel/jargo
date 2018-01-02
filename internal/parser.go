package internal

import (
	"reflect"
	"errors"
	"crushedpixel.net/jargo/internal/parser"
	"fmt"
	"github.com/c9s/inflect"
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
	errUnannotatedIdField        = errors.New("id field is missing jargo annotation")
	errInvalidMemberName         = errors.New(`member name has to adhere to the jsonapi specification and not include characters marked as "not recommended"`)
	errInvalidTableName          = errors.New("table name may only consist of [0-9,a-z,A-Z$_]")
	errInvalidTableAlias         = errors.New("alias may only consist of [0-9,a-z,A-Z$_]")
	errAliasEqualsTableName      = errors.New("alias may not be equal to table name")
	errInvalidHasType            = errors.New("has relation field has to be a struct pointer or slice of struct pointers")
	errInvalidMany2ManyType      = errors.New("many2many field has to be a slice of struct pointers")
	errMissingMany2ManyJoinTable = errors.New("missing many2many join table definition")
	errMultipleRelationTypes     = errors.New("multiple relation type options are not allowed")

	errDisallowedOption = func(option string) error {
		return errors.New(fmt.Sprintf(`option "%s" is not allowed here`, option))
	}
)

// parses a schema definition from a resource model type.
func (r Registry) newSchemaDefinition(t reflect.Type) (*schema, error) {
	if t.Kind() != reflect.Struct {
		return nil, errInvalidModelType
	}

	schema, err := parseSchema(t)
	if err != nil {
		return nil, err
	}

	// parse struct fields
	var jsonapiJoinFields []reflect.StructField
	var pgJoinFields []reflect.StructField
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		field, err := r.parseField(schema, &f)
		if err != nil {
			return nil, err
		}
		schema.fields = append(schema.fields, field)

		jf, err := field.jsonapiJoinFields()
		if err != nil {
			return nil, err
		}
		jsonapiJoinFields = append(jsonapiJoinFields, jf...)

		pf, err := field.pgJoinFields()
		if err != nil {
			return nil, err
		}
		pgJoinFields = append(pgJoinFields, pf...)
	}

	schema.joinJsonapiModelType = reflect.StructOf(jsonapiJoinFields)
	schema.joinPGModelType = reflect.StructOf(pgJoinFields)
	return schema, nil
}

func (r Registry) generateSchemaModels(schema *schema) error {
	var jsonapiFields []reflect.StructField
	var pgFields []reflect.StructField
	for _, f := range schema.fields {
		jf, err := f.jsonapiFields()
		if err != nil {
			return err
		}
		jsonapiFields = append(jsonapiFields, jf...)

		pf, err := f.pgFields()
		if err != nil {
			return err
		}
		pgFields = append(pgFields, pf...)
	}

	schema.jsonapiModelType = reflect.StructOf(jsonapiFields)
	schema.pgModelType = reflect.StructOf(pgFields)
	return nil
}

// parses a struct's id field, retrieving
// general schema information from the jargo struct tag.
func parseSchema(t reflect.Type) (*schema, error) {
	f, ok := t.FieldByName(idFieldName)
	if !ok {
		return nil, errMissingIdField
	}
	if f.Type != idFieldType {
		return nil, errInvalidIdType
	}

	// parse jargo struct tag
	tag, ok := f.Tag.Lookup(jargoFieldTag)
	if !ok {
		return nil, errUnannotatedIdField
	}

	// parse schema name, sql table and sql alias
	// from struct tag.
	// default sql alias is the snake_cased struct name.
	// default name and sql table is the pluralized version
	// of the alias.
	// "UserProfile" => "user_profile", "user_profiles"
	singleName := inflect.Underscore(t.Name())
	defaultName := inflect.Pluralize(singleName)
	parsed := parser.ParseJargoTagDefaultName(tag, defaultName)

	schema := &schema{
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
			return nil, errDisallowedOption(option)
		}
	}

	// validate member name
	if !isValidJsonapiMemberName(schema.name) {
		return nil, errInvalidMemberName
	}

	// validate table name
	if !isValidSQLName(schema.table) {
		return nil, errInvalidTableName
	}

	// validate alias
	if !isValidSQLName(schema.alias) {
		return nil, errInvalidTableAlias
	}

	// ensure alias is not the same as table name
	if schema.alias == schema.table {
		return nil, errAliasEqualsTableName
	}

	return schema, nil
}

// parses a struct field into a schema field.
// returns nil, nil for non-attribute fields.
func (r Registry) parseField(schema *schema, f *reflect.StructField) (field, error) {
	if f.Name == idFieldName {
		return newIdField(schema)
	}

	// determine field type from jargo tag
	parsed := parser.ParseJargoTag(f.Tag.Get(jargoFieldTag))

	typ := attribute
	var val string
	for option, value := range parsed.Options {
		switch option {
		case optionHas:
			if typ != attribute {
				return nil, errMultipleRelationTypes
			}
			typ = has
			val = value
		case optionBelongsTo:
			if typ != attribute {
				return nil, errMultipleRelationTypes
			}
			typ = belongsTo
		case optionMany2Many:
			if typ != attribute {
				return nil, errMultipleRelationTypes
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
		return nil, nil // TODO
	default:
		return newAttrField(schema, f)
	}
}
