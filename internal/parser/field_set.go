package parser

import (
	"net/url"
	"github.com/goji/param"
)

type fieldParameters struct {
	Fields  map[string]string `param:"fields"`
}

func ParseFieldParameters(query url.Values) (map[string]string, error) {
	p := new(fieldParameters)
	err := param.Parse(query, p)
	if err != nil {
		return nil, err
	}
	return p.Fields, nil
}
