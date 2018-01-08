package jargo

import "github.com/go-pg/pg/types"

// escapes a go-pg column string according to postgres rules.
// example: user.id => "user"."id"
func escapePGColumn(field string) string {
	var b []byte
	b = types.AppendField(b, field, 1)
	return string(b)
}
