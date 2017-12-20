package parser

import (
	"net/url"
	"github.com/goji/param"
)

type pageParameters struct {
	Page map[string]string `param:"page"`
}

func ParsePageParameters(query url.Values) (map[string]string, error) {
	p := new(pageParameters)
	err := param.Parse(query, p)
	if err != nil {
		return nil, err
	}
	return p.Page, nil
}
