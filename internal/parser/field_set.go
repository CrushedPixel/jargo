package parser

import (
	"net/url"
	"regexp"
	"strings"
)

var fieldsParamRegex = regexp.MustCompile(`^fields\[([^][]+)]$`)

// TODO: unit test and generify code with pagination parsing
func ParseFieldParameters(query url.Values) map[string][]string {
	fields := make(map[string][]string)
	for k, v := range query {
		res := fieldsParamRegex.FindAllStringSubmatch(k, -1)
		if res == nil {
			continue
		}
		groups := res[0]
		field := groups[1]

		// parse string-separated values
		values := make([]string, 0)
		for _, val := range v {
			values = append(values, strings.Split(val, ",")...)
		}

		fields[field] = values
	}

	return fields
}
