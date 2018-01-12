package parser

import (
	"net/url"
	"strings"
)

func ParseSortParameters(query url.Values) map[string]bool {
	values := make(map[string]bool)
	if sort, ok := query["sort"]; ok {
		for _, v := range sort {
			for _, fieldName := range strings.Split(v, ",") {
				// skip empty sort parameters
				if len(fieldName) < 1 {
					continue
				}

				asc := true
				if fieldName[0] == '-' {
					// if parameter is prefixed with a hyphen,
					// order is descending
					asc = false
					fieldName = fieldName[1:]
				}

				values[fieldName] = asc
			}
		}
	}

	return values
}
