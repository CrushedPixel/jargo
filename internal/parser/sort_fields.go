package parser

import (
	"net/url"
	"strings"
)

func ParseSortParameters(query url.Values) []string {
	values := make([]string, 0)
	if sort, ok := query["sort"]; ok {
		for _, v := range sort {
			values = append(values, strings.Split(v, ",")...)
		}
	}

	return values
}
