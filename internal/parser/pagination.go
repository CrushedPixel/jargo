package parser

import (
	"net/url"
	"regexp"
)

var pageParamRegex = regexp.MustCompile(`^page\[([^][]+)]$`)

// TODO: unit test and generify code with field set parsing
func ParsePageParameters(query url.Values) map[string]string {
	fields := make(map[string]string)
	for k, v := range query {
		res := pageParamRegex.FindAllStringSubmatch(k, -1)
		if res == nil {
			continue
		}
		groups := res[0]
		fields[groups[1]] = v[0]
	}

	return fields
}
