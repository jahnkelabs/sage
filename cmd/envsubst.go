package cmd

import (
	"fmt"
	"os"
	"strings"
)

// expandEnv replaces $VAR and ${VAR} in s using os.Getenv.
// Missing variables become empty strings and optionally emit warnings via warn.
func expandEnv(s string, warn func(string)) string {
	out := strings.Builder{}
	i := 0
	for i < len(s) {
		if s[i] != '$' {
			out.WriteByte(s[i])
			i++
			continue
		}
		if i+1 >= len(s) {
			out.WriteByte('$')
			i++
			continue
		}
		j := i + 1
		var name string
		if s[j] == '{' {
			end := strings.IndexByte(s[j+1:], '}')
			if end == -1 {
				out.WriteByte('$')
				i++
				continue
			}
			name = s[j+1 : j+1+end]
			i = j + 1 + end + 1
		} else {
			start := j
			for j < len(s) && (isAlphaNumUnderscore(s[j])) {
				j++
			}
			if start == j {
				out.WriteByte('$')
				i++
				continue
			}
			name = s[start:j]
			i = j
		}
		val, ok := os.LookupEnv(name)
		if !ok && warn != nil {
			warn(fmt.Sprintf("environment variable %s is not set; substituting empty string", name))
		}
		out.WriteString(val)
	}
	return out.String()
}

func isAlphaNumUnderscore(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}
