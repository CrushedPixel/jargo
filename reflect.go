package jargo

import (
	"reflect"
	"strings"
	"github.com/go-pg/pg/orm"
	"github.com/go-pg/pg/types"
)

const (
	annotationJSONAPI   = "jsonapi"
	annotationAttribute = "attr"

	annotationSeparator = ","

	annotationJargo  = "jargo"
	annotationFilter = "filter"
	annotationSort   = "sort"
)

type jargoAnnotations struct {
	filter bool // whether index results may be filtered by the attribute
	sort   bool // whether index results may be sorted by the attribute
}

// get sql column name from key by looking for
// attr name in model's jsonapi tag
func sqlColumnForAttrName(model interface{}, name string) (*types.Q, *jargoAnnotations) {
	t := reflect.TypeOf(model).Elem()
	field := fieldWithJsonApiAttrName(t, name)
	if field == nil {
		return nil, nil
	}

	// get sql column for field
	for _, f := range orm.Tables.Get(t).Fields {
		if f.GoName == field.Name {
			// found sql column for field
			a := parseJargoAnnotations(field)
			return &f.Column, a
		}
	}

	return nil, nil
}

func fieldWithJsonApiAttrName(t reflect.Type, name string) *reflect.StructField {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// parse jsonapi tag to find field with attribute name
		jsonTag := field.Tag.Get(annotationJSONAPI)
		if jsonTag != "" {
			args := strings.Split(jsonTag, annotationSeparator)
			if args[0] == annotationAttribute && len(args) > 1 {
				if name == args[1] {
					return &field
				}
			}
		}
	}

	return nil
}

func parseJargoAnnotations(field *reflect.StructField) *jargoAnnotations {
	var filter, sort bool

	val, ok := field.Tag.Lookup(annotationJargo)
	if !ok {
		filter = true
		sort = true
	} else {
		spl := strings.Split(val, annotationSeparator)
		for _, s := range spl {
			switch s {
			case annotationFilter:
				filter = true
			case annotationSort:
				sort = true
			default:
				// TODO: error handling?
			}
		}
	}

	return &jargoAnnotations{
		filter: filter,
		sort:   sort,
	}
}
