package resource

import "regexp"

var sqlNameRegex = regexp.MustCompile(`^[0-9a-zA-Z$_]+$`)
var memberNameRegex = regexp.MustCompile(`^[[:alnum:]]([a-zA-Z0-9\-_]*[[:alnum:]])?$`)

func pluralize(val string) string {
	l := len(val)
	if l == 0 || val[l - 1] == 's' {
		return val
	}

	return val + "s"
}

func isValidJsonapiMemberName(val string) bool {
	return memberNameRegex.MatchString(val)
}

func isValidSQLName(val string) bool {
	return sqlNameRegex.MatchString(val)
}