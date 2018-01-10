package parser

import (
	"net/url"
	"regexp"
	"strings"
)

var filterParamRegex = regexp.MustCompile(`^filter\[([^][]+)](?:\[([^][]+)])?$`)

func ParseFilterParameters(query url.Values) map[string]map[string][]string {
	// map[field]map[operator][]values
	filters := make(map[string]map[string][]string)
	for k, v := range query {
		res := filterParamRegex.FindAllStringSubmatch(k, -1)
		if res == nil {
			continue
		}

		// there's at most one match per key
		groups := res[0]

		field := groups[1]
		op := groups[2]
		if op == "" {
			op = "EQ"
		}
		op = strings.ToUpper(op)

		// add values to operations
		operations := filters[field]
		if operations == nil {
			operations = make(map[string][]string)
		}

		values := make([]string, 0)
		for _, val := range v {
			values = append(values, strings.Split(val, ",")...)
		}
		operations[op] = values

		filters[field] = operations
	}

	return filters
}
