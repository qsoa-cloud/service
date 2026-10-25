package errtext

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"syscall"
	"testing"
)

func TestText(t *testing.T) {
	errX := errors.New("x failed")
	addr := &net.UnixAddr{Name: "/hunter2.sock", Net: "unix"}
	for _, c := range []struct {
		err  error
		want string
	}{
		{fmt.Errorf("'hunter2': %w", context.Canceled), "context canceled"},
		{fmt.Errorf("'hunter2': %w", context.DeadlineExceeded), "context deadline exceeded"},
		{fmt.Errorf("'hunter2': %w", io.EOF), "EOF"},
		{fmt.Errorf("'hunter2': %w", io.ErrUnexpectedEOF), "unexpected EOF"},
		{fmt.Errorf("'hunter2': %w", sql.ErrNoRows), "sql: no rows in result set"},
		{fmt.Errorf("'hunter2': %w", errX), "x failed"},
		{errors.Join(errX, io.EOF), "EOF"},
		{&net.OpError{Op: "dial", Net: "unix", Addr: addr, Err: os.NewSyscallError("connect", syscall.ECONNREFUSED)}, "dial: connection refused"},
		{fmt.Errorf("'hunter2': %w", &net.OpError{Op: "write", Net: "unix", Addr: addr, Err: syscall.EPIPE}), "write: broken pipe"},
		{&net.OpError{Op: "read", Net: "unix", Addr: addr, Err: errors.New("hunter2")}, "read"},
		{errors.New("hunter2"), "*errors.errorString"},
		{fmt.Errorf("'hunter2': %w", errors.New("hunter2")), "*fmt.wrapError"},
	} {
		if got := Text(c.err, []error{errX}); got != c.want || strings.Contains(got, "hunter2") {
			t.Errorf("Text(%q) = %q, want %q", c.err, got, c.want)
		}
	}
}
