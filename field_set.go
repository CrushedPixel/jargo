package jargo

import (
	"github.com/crushedpixel/jargo/internal"
	"github.com/go-pg/pg/orm"
	"github.com/google/jsonapi"
)

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
