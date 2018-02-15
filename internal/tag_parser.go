package internal

import (
	"strings"
)

type JargoTag struct {
	Name    string
	Options map[string]string
}

func parseJargoTagDefaultName(tag string, defaultName string) *JargoTag {
	parsed := parseJargoTag(tag)
	if parsed.Name == "" {
		parsed.Name = defaultName
	}

	return parsed
}

func parseJargoTag(tag string) *JargoTag {
	spl := strings.Split(tag, ",")
	parsed := &JargoTag{
		Name:    spl[0],
		Options: make(map[string]string),
	}

	// parse options
	if len(spl) > 1 {
		for _, str := range spl[1:] {
			kv := strings.SplitN(str, ":", 2)
			k := kv[0]
			var v string
			if len(kv) < 2 {
				v = ""
			} else {
				v = kv[1]
			}

			parsed.Options[k] = v
		}
	}

	return parsed
}
