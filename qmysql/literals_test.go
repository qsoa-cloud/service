package qmysql

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestStripLiterals(t *testing.T) {
	for in, want := range map[string]string{
		"SELECT * FROM t WHERE a = 'x' AND b = 42":                                          "SELECT * FROM t WHERE a = ? AND b = ?",
		`INSERT INTO t2 VALUES ('it''s', "q\"", 3.5e2, 0xFF, X'0a', -7)`:                    "INSERT INTO t2 VALUES (?, ?, ?, ?, ?, -?)",
		"SELECT col1 FROM `tab1` WHERE c2 = ?":                                              "SELECT col1 FROM `tab1` WHERE c2 = ?",
		"UPDATE users SET password = 'hunter2' WHERE id = 7":                                "UPDATE users SET password = ? WHERE id = ?",
		`SELECT 'a\'b', 'unterminated`:                                                      "SELECT ?, ?",
		"SELECT a -- it's\nWHERE p = 'hunter2'":                                             "SELECT a -- it's\nWHERE p = ?",
		"SELECT a # it's\nWHERE p = 'hunter2'":                                              "SELECT a # it's\nWHERE p = ?",
		"SELECT a /* it's */ WHERE p = 'hunter2'":                                           "SELECT a /* it's */ WHERE p = ?",
		"SELECT a /* it's":                                                                  "SELECT a /* it's",
		"SELECT 5--1, 'x'":                                                                  "SELECT ?--?, ?",
		"SELECT /*+ MAX_EXECUTION_TIME(1000) */ a FROM t WHERE p = /*!80000 'hunter2' */ 1": "SELECT /*+ MAX_EXECUTION_TIME(?) */ a FROM t WHERE p = /*!? ? */ ?",
	} {
		if got := stripLiterals(in); got != want {
			t.Errorf("%q: %q, want %q", in, got, want)
		}
	}
	long := "SELECT " + strings.Repeat("é", 1000)
	if got := stripLiterals(long); len(got) > 1024 || !strings.HasPrefix(long, got) || !utf8.ValidString(got) {
		t.Errorf("long: %d bytes", len(got))
	}
}
