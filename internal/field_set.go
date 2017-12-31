package internal

import (
	"github.com/go-pg/pg/orm"
	"crushedpixel.net/jargo/api"
	"github.com/google/jsonapi"
	"net/url"
	"crushedpixel.net/jargo/internal/parser"
	"fmt"
	"errors"
)

type fieldSet struct {
	resource api.Resource
	fields   []field
}

func allFields(r *resource) api.FieldSet {
	fs := &fieldSet{
		resource: r,
	}
	copy(fs.fields, r.fields)
	return fs
}

func (fs *fieldSet) ApplyToQuery(q *orm.Query) {
	for _, f := range fs.fields {
		q.Column(f.pgColumn())
	}
}

func (fs *fieldSet) ApplyToJsonapiNode(node *jsonapi.Node) {
	if node.Type == fs.resource.Name() {
		fs.applyToPropertyMap(node.Attributes)
		fs.applyToPropertyMap(node.Relationships)
	}
}

func (fs *fieldSet) applyToPropertyMap(m map[string]interface{}) {
	for key := range m {
		found := false

		for _, field := range fs.fields {
			if field.jsonapiName() == key {
				found = true
				break
			}
		}

		if !found {
			delete(m, key)
		}
	}
}

func parseFieldSet(r *resource, query url.Values) (api.FieldSet, error) {
	parsed := parser.ParseFieldParameters(query)

	var resourceFields []field
	if fields, ok := parsed[r.Name()]; ok {
		for _, fieldName := range fields {
			// find resource field with matching jsonapi name
			var field field
			for _, rf := range r.fields {
				if rf.jsonapiName() == fieldName {
					field = rf
					break
				}
			}
			if field == nil {
				return nil, errors.New(fmt.Sprintf(`unknown field parameter: "%s"`, fieldName))
			}
			resourceFields = append(resourceFields, field)
		}

		removeDuplicateFields(resourceFields)
		return &fieldSet{
			resource: r,
			fields:   resourceFields,
		}, nil
	}

	return allFields(r), nil
}

func removeDuplicateFields(fields []field) {
	found := make(map[field]bool)
	j := 0
	for i, x := range fields {
		if !found[x] {
			found[x] = true
			fields[j] = fields[i]
			j++
		}
	}
	fields = fields[:j]
}
