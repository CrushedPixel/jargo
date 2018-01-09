package parser

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var operators = []string{"EQ", "NOT", "LIKE", "LT", "LTE", "GT", "GTE"}

func errInvalidOperator(op string) error {
	return errors.New(fmt.Sprintf(`unknown filter operator: %s`, op))
}

var filterParamRegex = regexp.MustCompile(`^filter\[([^][]+)](?:\[([^][]+)])?$`)

func ParseFilterParameters(query url.Values) (map[string]map[string][]string, error) {
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

		// validate operator
		var operator string
		for _, o := range operators {
			if strings.ToUpper(op) == o {
				operator = o
				break
			}
		}

		if operator == "" {
			return nil, errInvalidOperator(op)
		}

		// add values to operations
		operations := filters[field]
		if operations == nil {
			operations = make(map[string][]string)
		}

		values := make([]string, 0)
		for _, val := range v {
			values = append(values, strings.Split(val, ",")...)
		}
		operations[operator] = values

		filters[field] = operations
	}

	return filters, nil
}
