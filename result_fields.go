package jargo

import (
	"strings"
	"github.com/google/jsonapi"
)

type ResultFields map[string][]string

func (r ResultFields) ApplyToNode(node *jsonapi.Node) {
	if fields, ok := r[node.Type]; ok {
		// if sparse fieldset is requested for type,
		// remove fields that are not wanted
		r.applyToFieldset(node.Attributes, fields)
		r.applyToFieldset(node.Relationships, fields)
	}
}

func (r ResultFields) applyToFieldset(fieldset map[string]interface{}, fields []string) {
	if fieldset != nil {
		for field, value := range fieldset {
			if !containsValue(fields, field) {
				delete(fieldset, field)
			} else {
				// check if value is a node,
				// applying the constraints recursively
				n2, ok := value.(*jsonapi.Node)
				if ok {
					r.ApplyToNode(n2)
				}
			}
		}
	}
}

func parseFieldParameters(values map[string]string) (ResultFields, error) {
	fieldSets := make(ResultFields)

	for k, v := range values {
		values := strings.Split(v, ",")
		fieldSets[k] = values
	}

	return fieldSets, nil
}
