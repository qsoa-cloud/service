package qmysql

import (
	"strings"
	"unicode/utf8"
)

const maxSQL = 1024

func stripLiterals(sql string) string {
	var b strings.Builder
	b.Grow(min(len(sql), maxSQL))
	from := 0
	for i := 0; i < len(sql) && i-from <= maxSQL-b.Len(); {
		n, lit := token(sql, i)
		if lit {
			if !put(&b, sql[from:i]) || !put(&b, "?") {
				return b.String()
			}
			from = i + n
		}
		i += n
	}
	put(&b, sql[from:])
	return b.String()
}

func put(b *strings.Builder, s string) bool {
	n := maxSQL - b.Len()
	if len(s) <= n {
		b.WriteString(s)
		return true
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	b.WriteString(s[:n])
	return false
}

func token(s string, i int) (int, bool) {
	c := s[i]
	switch {
	case c == '\'' || c == '"':
		return quoted(s[i:]), true
	case c == '`':
		return 1 + through(s[i+1:], "`"), false
	case c == '#' || c == '-' && i+2 < len(s) && s[i+1] == '-' && s[i+2] <= ' ':
		return through(s[i:], "\n"), false
	case c == '/' && i+2 < len(s) && s[i+1] == '*' && s[i+2] != '!' && s[i+2] != '+':
		return 2 + through(s[i+2:], "*/"), false
	case i > 0 && word(s[i-1]):
		return 1, false
	case (c|0x20 == 'x' || c|0x20 == 'b') && i+1 < len(s) && s[i+1] == '\'':
		return 1 + quoted(s[i+1:]), true
	case c >= '0' && c <= '9':
		n := 1
		for n < len(s)-i && (word(s[i+n]) || s[i+n] == '.' || (s[i+n] == '+' || s[i+n] == '-') && s[i+n-1]|0x20 == 'e') {
			n++
		}
		return n, true
	}
	return 1, false
}

func through(s, end string) int {
	if j := strings.Index(s, end); j >= 0 {
		return j + len(end)
	}
	return len(s)
}

func quoted(s string) int {
	for i := 1; i < len(s); i++ {
		switch s[i] {
		case '\\':
			i++
		case s[0]:
			if i+1 == len(s) || s[i+1] != s[0] {
				return i + 1
			}
			i++
		}
	}
	return len(s)
}

func word(c byte) bool {
	return c >= 0x80 || c == '_' || c == '$' || c >= '0' && c <= '9' || c|0x20 >= 'a' && c|0x20 <= 'z'
}
