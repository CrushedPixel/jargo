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
func difference(a, b []int64) []int64 {
	mb := make(map[int64]bool)
	for _, x := range b {
		mb[x] = true
	}
	var ab []int64
	for _, x := range a {
		if _, ok := mb[x]; !ok {
			ab = append(ab, x)
		}
	}
	return ab
}

// normalizeNamespace ensures that namespace starts and ends with a slash.
func normalizeNamespace(namespace string) string {
	// prepend slash to namespace
	if len(namespace) < 1 || namespace[0] != '/' {
		namespace = "/" + namespace
	}
	// remove trailing slash from namespace
	if len(namespace) > 1 && namespace[len(namespace)-1] != '/' {
		namespace = namespace + "/"
	}

	return namespace
}
