package internal

import (
	"reflect"
	"crushedpixel.net/jargo/internal/parser"
	"github.com/c9s/inflect"
)

type baseField struct {
	schema *schema

	fieldName string
	fieldType reflect.Type

	name string

	jargoWritable   bool
	jargoSortable   bool
	jargoFilterable bool

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

func (f *baseField) writable() bool {
	return f.jargoWritable
}

func (f *baseField) sortable() bool {
	return f.jargoSortable
}

func (f *baseField) filterable() bool {
	return f.jargoFilterable
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
	// determine default jsonapi member name.
	// defaults to dasherized struct field name.
	defaultName := inflect.Dasherize(f.Name)
	parsed := parser.ParseJargoTagDefaultName(f.Tag.Get(jargoFieldTag), defaultName)

	field := &baseField{
		schema:          schema,
		fieldName:       f.Name,
		fieldType:       f.Type,
		name:            parsed.Name,
		jargoWritable:   true,
		jargoSortable:   true,
		jargoFilterable: true,
		jsonapiExported: parsed.Name != unexportedFieldName,
	}

	// parse options
	for option, value := range parsed.Options {
		switch option {
		case optionReadonly:
			b, err := parseBoolOption(value)
			if err != nil {
				return nil, err
			}
			field.jargoWritable = !b
		case optionSort:
			b, err := parseBoolOption(value)
			if err != nil {
				return nil, err
			}
			field.jargoSortable = b
		case optionFilter:
			b, err := parseBoolOption(value)
			if err != nil {
				return nil, err
			}
			field.jargoFilterable = b
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
