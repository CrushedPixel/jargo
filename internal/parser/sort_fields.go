package parser

import (
	"net/url"
)

func ParseSortParameters(query url.Values) string {
	return query.Get("sort")
}
