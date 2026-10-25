package sqltext

import "testing"

func TestStripDialects(t *testing.T) {
	for in, want := range map[string]string{
		"SELECT $$hunter2$$":                                 "SELECT ?",
		"SELECT $x$it's 'a' $$ hunter2$x$, 'b'":              "SELECT ?, ?",
		"SELECT $$hunter2":                                   "SELECT ?",
		"SELECT x$$hunter2$$":                                "SELECT x?",
		"SELECT a --don't\nWHERE p = 'hunter2'":              "SELECT a --don't\nWHERE p = ?",
		"SELECT 1 /*+ it's */ WHERE p = 'hunter2'":           "SELECT ? /*+ it's */ WHERE p = ?",
		"SELECT 1 /*! it's */ WHERE p = 'hunter2'":           "SELECT ? /*! it's */ WHERE p = ?",
		"SELECT a # it's\nWHERE p = 'hunter2'":               "SELECT a # it's\nWHERE p = ?",
		"SELECT a #!it's\nWHERE p = 'hunter2'":               "SELECT a #!it's\nWHERE p = ?",
		"SELECT a #x 'hunter2'":                              "SELECT a #x ?",
		"SELECT 1 AS `a\\`b`, 'hunter2'":                     "SELECT ? AS `a\\`b`, ?",
		"SELECT 1 AS `a``b`, 'hunter2' FROM `t`":             "SELECT ? AS `a``b`, ? FROM `t`",
		"SELECT a FROM t WHERE b = 'x' AND c = 4.5e3":        "SELECT a FROM t WHERE b = ? AND c = ?",
		"SELECT a // it's\nWHERE p = 'hunter2'":              "SELECT a // it's\nWHERE p = ?",
		"SELECT 1 /* a /* b */ don't */ WHERE p = 'hunter2'": "SELECT ? /* a /* b */ don't */ WHERE p = ?",
		"SELECT 1 /*/ it's */ WHERE p = 'hunter2'":           "SELECT ? /*/ it's */ WHERE p = ?",
		"SELECT 1 /* a /* b */ it's":                         "SELECT ? /* a /* b */ it's",
		`SELECT "col" FROM t WHERE p = 'x'`:                  `SELECT "col" FROM t WHERE p = ?`,
	} {
		if got := Strip(in, ClickHouse); got != want {
			t.Errorf("ClickHouse %q: %q, want %q", in, got, want)
		}
	}
	for _, in := range []string{"SELECT $$hunter2$$", "SELECT $x$hunter2$x$", "SELECT x$$hunter2$$"} {
		if got := Strip(in, MySQL); got != in {
			t.Errorf("MySQL %q: %q", in, got)
		}
	}
	if got := Strip(`SELECT "col" FROM t WHERE p = 'x'`, MySQL); got != "SELECT ? FROM t WHERE p = ?" {
		t.Errorf("MySQL double quotes: %q", got)
	}
}
