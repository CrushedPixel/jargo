package jargo

import (
	"strings"
)

type ResultFields map[string][]string

func parseFieldParameters(values map[string]string) (*ResultFields, error) {
	fieldSets := make(ResultFields)

	for k, v := range values {
		values := strings.Split(v, ",")
		fieldSets[k] = values
	}

	return &fieldSets, nil
}
