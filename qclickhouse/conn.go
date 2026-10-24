package qclickhouse

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	chproto "github.com/ClickHouse/ch-go/proto"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"

	"gopkg.qsoa.cloud/service/internal/sqltext"
)

type conn struct {
	driver.Conn
	db string
}

func (c *conn) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "clickhouse query")
	rows, err := c.Conn.Query(ctx, query, args...)
	tag(span, c.db, query, err)
	if err != nil {
		span.Finish()
		return nil, err
	}
	return &tracedRows{Rows: rows, span: span}, nil
}

func (c *conn) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	span, ctx := opentracing.StartSpanFromContext(ctx, "clickhouse query row")
	row := c.Conn.QueryRow(ctx, query, args...)
	tag(span, c.db, query, row.Err())
	if row.Err() != nil {
		span.Finish()
		return row
	}
	return &tracedRow{Row: row, span: span}
}

func (c *conn) Select(ctx context.Context, dest any, query string, args ...any) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "clickhouse select")
	err := c.Conn.Select(ctx, dest, query, args...)
	end(span, c.db, query, err)
	return err
}

func (c *conn) Exec(ctx context.Context, query string, args ...any) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "clickhouse exec")
	err := c.Conn.Exec(ctx, query, args...)
	end(span, c.db, query, err)
	return err
}

func (c *conn) AsyncInsert(ctx context.Context, query string, wait bool, args ...any) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "clickhouse async insert")
	err := c.Conn.AsyncInsert(ctx, query, wait, args...)
	end(span, c.db, query, err)
	return err
}

func (c *conn) PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "clickhouse batch")
	b, err := c.Conn.PrepareBatch(ctx, query, opts...)
	tag(span, c.db, query, err)
	if err != nil {
		span.Finish()
		return nil, err
	}
	return &tracedBatch{Batch: b, span: span}, nil
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
	span.SetTag("sql", sqltext.Strip(query, true))
	fail(span, err)
}

func fail(span opentracing.Span, err error) {
	if err != nil {
		span.SetTag("error", nil)
		span.LogFields(log.String("error.object", errText(err)))
	}
}

func errText(err error) string {
	var e *clickhouse.Exception
	if errors.As(err, &e) {
		return "Code " + strconv.Itoa(int(e.Code)) + " (" + chproto.Error(e.Code).String() + ")"
	}
	return sqltext.Strip(err.Error(), true)
}
