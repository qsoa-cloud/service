package sqltext

import "testing"

func TestStripHeredoc(t *testing.T) {
	for in, want := range map[string]string{
		"SELECT $$hunter2$$":                    "SELECT ?",
		"SELECT $x$it's 'a' $$ hunter2$x$, 'b'": "SELECT ?, ?",
		"SELECT $$hunter2":                      "SELECT ?",
		"SELECT a$b$c, $1 FROM t":               "SELECT a$b$c, $1 FROM t",
	} {
		if got := Strip(in, true); got != want {
			t.Errorf("%q: %q, want %q", in, got, want)
		}
	}
	for _, in := range []string{"SELECT $$hunter2$$", "SELECT $x$hunter2$x$"} {
		if got := Strip(in, false); got != in {
			t.Errorf("without heredoc %q: %q", in, got)
		}
	}
}
