package jargo

import (
	"fmt"
	"github.com/crushedpixel/jargo/internal"
	"github.com/go-pg/pg/orm"
	"github.com/google/jsonapi"
)

// FieldSet contains information about
// which fields are
type FieldSet struct {
	resource *Resource
	fields   []internal.SchemaField
}

func newFieldSet(r *Resource, fields []internal.SchemaField) *FieldSet {
	removeDuplicateFields(fields)
	return &FieldSet{
		resource: r,
		fields:   fields,
	}
}

func (fs *FieldSet) applyToQuery(q *orm.Query) {
	for _, f := range fs.fields {
		q.Column(f.PGSelectColumn())
	}
}

func (fs *FieldSet) applyToJsonapiNode(node *jsonapi.Node) {
	if node.Type == fs.resource.JSONAPIName() {
		fs.applyToPropertyMap(node.Attributes)
		fs.applyToPropertyMap(node.Relationships)
	}
}

func (fs *FieldSet) applyToPropertyMap(m map[string]interface{}) {
	for key := range m {
		found := false

		for _, field := range fs.fields {
			if field.JSONAPIName() == key {
				found = true
				break
			}
		}

		if !found {
			delete(m, key)
		}
	}
}

func removeDuplicateFields(fields []internal.SchemaField) {
	found := make(map[internal.SchemaField]bool)
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

// FieldNames returns the JSON API Member names
// of the FieldSet's fields.
func (fs *FieldSet) FieldNames() []string {
	var names []string
	for _, f := range fs.fields {
		names = append(names, f.JSONAPIName())
	}
	return names
}

// Without returns a copy of the FieldSet instance
// not containing fields with the JSON API Member names passed.
func (fs *FieldSet) Without(names ...string) *FieldSet {
	f := &FieldSet{
		resource: fs.resource,
	}

	for _, field := range fs.fields {
		excluded := false
		for _, name := range names {
			if field.JSONAPIName() == name {
				excluded = true
			}
			break
		}

		if !excluded {
			f.fields = append(f.fields, field)
		}
	}

	return f
}

// ParseFieldSet creates a FieldSet instance
// for the given map of field parameters.
// These parameters can be created manually
// or extracted from an URL's query parameters
// using ParseFieldParameters.
//
// Returns ErrInvalidQueryParams when encountering invalid query values.
func (r *Resource) ParseFieldSet(parsed map[string][]string) (*FieldSet, error) {
	var schemaFields []internal.SchemaField
	// check if user specified to filter this resource's fields
	if fields, ok := parsed[r.JSONAPIName()]; ok {
		// always include the id field,
		// so it gets fetched from the database
		fields = append(fields, internal.IdFieldJsonapiName)
		for _, fieldName := range fields {
			// find resource field with matching jsonapi name
			var field internal.SchemaField
			for _, f := range r.schema.Fields() {
				if f.JSONAPIName() == fieldName {
					field = f
					break
				}
			}
			if field == nil {
				return nil, ErrInvalidQueryParams(fmt.Sprintf(`unknown field parameter: "%s"`, fieldName))
			}

			schemaFields = append(schemaFields, field)
		}
		return newFieldSet(r, schemaFields), nil
	} else {
		return r.allFields(), nil
	}
}
