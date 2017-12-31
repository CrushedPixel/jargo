package internal

import (
	"reflect"
	"github.com/iancoleman/strcase"
	"crushedpixel.net/jargo/internal/parser"
)

type baseField struct {
	schema *schema

	fieldName string
	fieldType reflect.Type

	name string // jsonapi attribute name

	jsonapiExported  bool
	jsonapiOmitempty bool

	sqlNotnull bool
	sqlUnique  bool

	jsonapiF []reflect.StructField
	pgF      []reflect.StructField
}

func (f *baseField) jsonapiName() string {
	return f.name
}

func (f *baseField) jsonapiFields() ([]reflect.StructField, error) {
	return f.jsonapiF, nil
}

func (f *baseField) jsonapiJoinFields() ([]reflect.StructField, error) {
	return f.jsonapiF, nil
}

func (f *baseField) pgFields() ([]reflect.StructField, error) {
	return f.pgF, nil
}

func (f *baseField) pgJoinFields() ([]reflect.StructField, error) {
	return f.pgF, nil
}

func newBaseField(schema *schema, f *reflect.StructField) (*baseField, error) {
	// determine jsonapi member name,
	// defaulting to snake_cased field name
	defaultName := strcase.ToSnake(f.Name)
	parsed := parser.ParseJargoTagDefaultName(f.Tag.Get(jargoFieldTag), defaultName)

	field := &baseField{
		schema:          schema,
		fieldName:       f.Name,
		fieldType:       f.Type,
		name:            parsed.Name,
		jsonapiExported: f.Name != unexportedFieldName,
	}

	// parse options
	for option, value := range parsed.Options {
		switch option {
		case optionOmitempty:
			if !field.jsonapiExported {
				return nil, errJsonapiOptionOnUnexportedField
			}
			b, err := parseBoolOption(value)
			if err != nil {
				return nil, err
			}
			field.jsonapiOmitempty = b
		case optionNotnull:
			b, err := parseBoolOption(value)
			if err != nil {
				return nil, err
			}
			field.sqlNotnull = b
		case optionUnique:
			b, err := parseBoolOption(value)
			if err != nil {
				return nil, err
			}
			field.sqlUnique = b
		}
	}

	return field, nil
}
