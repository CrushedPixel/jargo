package jargo

import "github.com/go-pg/pg/types"

// escapes a go-pg column string according to postgres rules.
// example: user.id => "user"."id"
func escapePGColumn(field string) string {
	var b []byte
	b = types.AppendField(b, field, 1)
	return string(b)
}

// difference returns the elements in a that aren't in b.
// Modified from https://stackoverflow.com/a/45428032/2733724
func difference(a, b []interface{}) []interface{} {
	mb := make(map[interface{}]bool)
	for _, x := range b {
		mb[x] = true
	}
	var ab []interface{}
	for _, x := range a {
		if _, ok := mb[x]; !ok {
			ab = append(ab, x)
		}
	}
	return ab
}

// NormalizeNamespace ensures that the namespace starts and ends with a slash.
func NormalizeNamespace(namespace string) string {
	// prepend slash to namespace
	if len(namespace) < 1 || namespace[0] != '/' {
		namespace = "/" + namespace
	}
	// add trailing slash to namespace
	if len(namespace) > 1 && namespace[len(namespace)-1] != '/' {
		namespace = namespace + "/"
	}

	return namespace
}

// NilNotFound returns ErrNotFound if res is nil.
// May be used to wrap calls to jargo.Query.Result() to
// avoid having to handle the nil case explicitly:
// res, err := jargo.NilNotFound(q.Result())
func NilNotFound(res interface{}, err error) (interface{}, error) {
	if res == nil {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}
