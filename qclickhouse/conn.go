package qclickhouse

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strconv"
	"time"

	chproto "github.com/ClickHouse/ch-go/proto"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/column"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"

	"gopkg.qsoa.cloud/service/internal/errtext"
	"gopkg.qsoa.cloud/service/internal/sqltext"
)

type conn struct {
	driver.Conn
	db string
}

func (c *conn) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "clickhouse query")
	rows, err := c.Conn.Query(driverCtx(ctx), query, args...)
	tag(span, c.db, query, err)
	if err != nil {
		span.Finish()
		return nil, err
	}
	return &tracedRows{Rows: rows, span: span}, nil
}

func (c *conn) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	span, ctx := opentracing.StartSpanFromContext(ctx, "clickhouse query row")
	row := c.Conn.QueryRow(driverCtx(ctx), query, args...)
	tag(span, c.db, query, row.Err())
	if row.Err() != nil {
		span.Finish()
		return row
	}
	return &tracedRow{Row: row, span: span}
}

func (c *conn) Select(ctx context.Context, dest any, query string, args ...any) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "clickhouse select")
	err := c.Conn.Select(driverCtx(ctx), dest, query, args...)
	end(span, c.db, query, err)
	return err
}

func (c *conn) Exec(ctx context.Context, query string, args ...any) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "clickhouse exec")
	err := c.Conn.Exec(driverCtx(ctx), query, args...)
	end(span, c.db, query, err)
	return err
}

func (c *conn) AsyncInsert(ctx context.Context, query string, wait bool, args ...any) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "clickhouse async insert")
	err := c.Conn.AsyncInsert(driverCtx(ctx), query, wait, args...)
	end(span, c.db, query, err)
	return err
}

func (c *conn) PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "clickhouse batch")
	b, err := c.Conn.PrepareBatch(driverCtx(ctx), query, opts...)
	tag(span, c.db, query, err)
	if err != nil {
		span.Finish()
		return nil, err
	}
	return &tracedBatch{Batch: b, span: span}, nil
}

type noDeadline struct{ context.Context }

func (noDeadline) Deadline() (time.Time, bool) { return time.Time{}, false }

func driverCtx(ctx context.Context) context.Context {
	if _, ok := ctx.Deadline(); ok {
		return noDeadline{ctx}
	}
	return ctx
}

type tracedRows struct {
	driver.Rows
	span opentracing.Span
	n    int
}

func (r *tracedRows) Next() bool {
	if r.Rows.Next() {
		r.n++
		return true
	}
	r.finish(r.Rows.Err())
	return false
}

func (r *tracedRows) Close() error {
	err := r.Rows.Close()
	r.finish(err)
	return err
}

func (r *tracedRows) finish(err error) {
	if r.span != nil {
		r.span.SetTag("rows", r.n)
		fail(r.span, err)
		r.span.Finish()
		r.span = nil
	}
}

type tracedRow struct {
	driver.Row
	span opentracing.Span
}

func (r *tracedRow) Scan(dest ...any) error    { return r.finish(r.Row.Scan(dest...)) }
func (r *tracedRow) ScanStruct(dest any) error { return r.finish(r.Row.ScanStruct(dest)) }

func (r *tracedRow) finish(err error) error {
	if r.span != nil {
		if err != sql.ErrNoRows {
			fail(r.span, err)
		}
		r.span.Finish()
		r.span = nil
	}
	return err
}

type tracedBatch struct {
	driver.Batch
	span opentracing.Span
	n    int
}

func (b *tracedBatch) Flush() error {
	n := b.Batch.Rows()
	err := b.Batch.Flush()
	b.n += n - b.Batch.Rows()
	return err
}

func (b *tracedBatch) Send() error {
	n := b.n + b.Batch.Rows()
	err := b.Batch.Send()
	if b.span != nil {
		b.span.SetTag("rows", n)
	}
	b.finish(err)
	return err
}

func (b *tracedBatch) Abort() error {
	err := b.Batch.Abort()
	b.finish(err)
	return err
}

func (b *tracedBatch) Close() error {
	err := b.Batch.Close()
	b.finish(err)
	return err
}

func (b *tracedBatch) finish(err error) {
	if b.span != nil {
		fail(b.span, err)
		b.span.Finish()
		b.span = nil
	}
}

func end(span opentracing.Span, db, query string, err error) {
	tag(span, db, query, err)
	span.Finish()
}

func tag(span opentracing.Span, db, query string, err error) {
	span.SetTag("db", db)
	span.SetTag("sql", sqltext.Strip(query, sqltext.ClickHouse))
	fail(span, err)
}

func fail(span opentracing.Span, err error) {
	if err != nil {
		span.SetTag("error", nil)
		span.LogFields(log.String("error.object", errText(err)))
	}
}

var ident = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

var constErrs = []error{
	clickhouse.ErrBatchAlreadySent, clickhouse.ErrBatchNotSent, clickhouse.ErrAcquireConnTimeout,
	clickhouse.ErrUnsupportedServerRevision, clickhouse.ErrBindMixedParamsFormats, clickhouse.ErrAcquireConnNoAddress,
	clickhouse.ErrServerUnexpectedData, clickhouse.ErrConnectionClosed, clickhouse.ErrClusterSecretRequiresName,
	clickhouse.ErrClusterSecretNeedsNative, clickhouse.ErrClusterSecretRequiresUsername, clickhouse.ErrClusterSecretWithJWT,
	clickhouse.ErrFormatNativeUnsupported, clickhouse.ErrInvalidValueInNamedDateValue, clickhouse.ErrUnsupportedQueryParameter,
	clickhouse.ErrInvalidTimezone,
}

func errText(err error) string {
	var (
		e    *clickhouse.Exception
		op   *clickhouse.OpError
		conv *column.ColumnConverterError
		ut   *column.UnsupportedColumnTypeError
	)
	switch {
	case errors.As(err, &e):
		return "Code " + strconv.Itoa(int(e.Code)) + " (" + chproto.Error(e.Code).String() + ")"
	case errors.Is(err, clickhouse.ErrBatchInvalid):
		return "batch is invalid"
	case errors.As(err, &op):
		s := errText(op.Err)
		if ident.MatchString(op.ColumnName) {
			s = op.ColumnName + ": " + s
		}
		if op.Op != "" {
			s = "clickhouse [" + op.Op + "]: " + s
		}
		return s
	case errors.As(err, &conv):
		return "converting " + conv.From + " to " + conv.To + " is unsupported"
	case errors.As(err, &ut):
		return "unsupported column type"
	}
	return errtext.Text(err, constErrs)
}
