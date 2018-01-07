package internal

import (
	"github.com/c9s/inflect"
	"github.com/crushedpixel/jargo/internal/parser"
	"reflect"
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

func (f *baseField) jsonapiFields() []reflect.StructField {
	return f.jsonapiF
}

func (f *baseField) jsonapiJoinFields() []reflect.StructField {
	return f.jsonapiF
}

func (f *baseField) pgFields() []reflect.StructField {
	return f.pgF
}

func (f *baseField) pgJoinFields() []reflect.StructField {
	return f.pgF
}

func newBaseField(schema *schema, f *reflect.StructField) *baseField {
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
			field.jargoWritable = !parseBoolOption(value)
		case optionSort:
			field.jargoSortable = parseBoolOption(value)
		case optionFilter:
			field.jargoFilterable = parseBoolOption(value)
		case optionOmitempty:
			if !field.jsonapiExported {
				panic(errJsonapiOptionOnUnexportedField)
			}
			field.jsonapiOmitempty = parseBoolOption(value)
		case optionNotnull:
			field.sqlNotnull = parseBoolOption(value)
		case optionUnique:
			field.sqlUnique = parseBoolOption(value)
		}
	}

	return field
}
