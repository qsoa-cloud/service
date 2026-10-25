package errtext

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
)

var common = []error{context.Canceled, context.DeadlineExceeded, io.EOF, io.ErrUnexpectedEOF, sql.ErrNoRows}

// Text returns the text of the first common or consts error that err matches, the Op and errno of a *net.OpError, or the type of err.
func Text(err error, consts []error) string {
	for _, l := range [2][]error{common, consts} {
		for _, c := range l {
			if errors.Is(err, c) {
				return c.Error()
			}
		}
	}
	var (
		ne    *net.OpError
		errno syscall.Errno
	)
	if errors.As(err, &ne) {
		if errors.As(ne.Err, &errno) {
			return ne.Op + ": " + errno.Error()
		}
		return ne.Op
	}
	return fmt.Sprintf("%T", err)
}
