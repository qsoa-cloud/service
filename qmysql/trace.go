package qmysql

import (
	"context"
	"database/sql/driver"
	"io"
	"reflect"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

type tracedConnector struct {
	base driver.Connector
	drv  driver.Driver
	db   string
}

func (c *tracedConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.base.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return &tracedConn{conn: conn, db: c.db}, nil
}

func (c *tracedConnector) Driver() driver.Driver { return c.drv }

type tracedConn struct {
	conn driver.Conn
	db   string
}

func (c *tracedConn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

func (c *tracedConn) Close() error { return c.conn.Close() }

func (c *tracedConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *tracedConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	b, ok := c.conn.(driver.ConnBeginTx)
	if !ok {
		return nil, driver.ErrSkip
	}
	span, sctx := opentracing.StartSpanFromContext(ctx, "mysql begin")
	tx, err := b.BeginTx(sctx, opts)
	end(span, c.db, "", err)
	if err != nil {
		return nil, err
	}
	return &tracedTx{tx: tx, db: c.db, ctx: ctx}, nil
}

func (c *tracedConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	p, ok := c.conn.(driver.ConnPrepareContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	span, ctx := opentracing.StartSpanFromContext(ctx, "mysql prepare")
	stmt, err := p.PrepareContext(ctx, query)
	end(span, c.db, query, err)
	if err != nil {
		return nil, err
	}
	return &tracedStmt{stmt: stmt, db: c.db, query: query}, nil
}

func (c *tracedConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	q, ok := c.conn.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	span, ctx := opentracing.StartSpanFromContext(ctx, "mysql query")
	rows, err := q.QueryContext(ctx, query, args)
	return traceRows(span, c.db, query, rows, err)
}

func (c *tracedConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	e, ok := c.conn.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	span, ctx := opentracing.StartSpanFromContext(ctx, "mysql exec")
	res, err := e.ExecContext(ctx, query, args)
	return traceExec(span, c.db, query, res, err)
}

func (c *tracedConn) Ping(ctx context.Context) error {
	if p, ok := c.conn.(driver.Pinger); ok {
		return p.Ping(ctx)
	}
	return nil
}

func (c *tracedConn) ResetSession(ctx context.Context) error {
	if r, ok := c.conn.(driver.SessionResetter); ok {
		return r.ResetSession(ctx)
	}
	return nil
}

func (c *tracedConn) IsValid() bool {
	v, ok := c.conn.(driver.Validator)
	return !ok || v.IsValid()
}

func (c *tracedConn) CheckNamedValue(nv *driver.NamedValue) error {
	if n, ok := c.conn.(driver.NamedValueChecker); ok {
		return n.CheckNamedValue(nv)
	}
	return driver.ErrSkip
}

type tracedStmt struct {
	stmt  driver.Stmt
	db    string
	query string
}

func (s *tracedStmt) Close() error  { return s.stmt.Close() }
func (s *tracedStmt) NumInput() int { return s.stmt.NumInput() }

func (s *tracedStmt) Exec(args []driver.Value) (driver.Result, error) { return s.stmt.Exec(args) }
func (s *tracedStmt) Query(args []driver.Value) (driver.Rows, error)  { return s.stmt.Query(args) }

func (s *tracedStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	q, ok := s.stmt.(driver.StmtQueryContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	span, ctx := opentracing.StartSpanFromContext(ctx, "mysql stmt query")
	rows, err := q.QueryContext(ctx, args)
	return traceRows(span, s.db, s.query, rows, err)
}

func (s *tracedStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	e, ok := s.stmt.(driver.StmtExecContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	span, ctx := opentracing.StartSpanFromContext(ctx, "mysql stmt exec")
	res, err := e.ExecContext(ctx, args)
	return traceExec(span, s.db, s.query, res, err)
}

func (s *tracedStmt) CheckNamedValue(nv *driver.NamedValue) error {
	if n, ok := s.stmt.(driver.NamedValueChecker); ok {
		return n.CheckNamedValue(nv)
	}
	return driver.ErrSkip
}

func (s *tracedStmt) ColumnConverter(idx int) driver.ValueConverter {
	if c, ok := s.stmt.(driver.ColumnConverter); ok {
		return c.ColumnConverter(idx)
	}
	return driver.DefaultParameterConverter
}

type tracedRows struct {
	rows driver.Rows
	span opentracing.Span
	n    int
}

func (r *tracedRows) Columns() []string { return r.rows.Columns() }

func (r *tracedRows) Next(dest []driver.Value) error {
	err := r.rows.Next(dest)
	if err == nil {
		r.n++
	} else if err != io.EOF {
		fail(r.span, err)
	}
	return err
}

func (r *tracedRows) Close() error {
	err := r.rows.Close()
	r.span.SetTag("rows", r.n)
	fail(r.span, err)
	r.span.Finish()
	return err
}

func (r *tracedRows) HasNextResultSet() bool {
	n, ok := r.rows.(driver.RowsNextResultSet)
	return ok && n.HasNextResultSet()
}

func (r *tracedRows) NextResultSet() error {
	if n, ok := r.rows.(driver.RowsNextResultSet); ok {
		return n.NextResultSet()
	}
	return io.EOF
}

func (r *tracedRows) ColumnTypeDatabaseTypeName(i int) string {
	if c, ok := r.rows.(driver.RowsColumnTypeDatabaseTypeName); ok {
		return c.ColumnTypeDatabaseTypeName(i)
	}
	return ""
}

func (r *tracedRows) ColumnTypeScanType(i int) reflect.Type {
	if c, ok := r.rows.(driver.RowsColumnTypeScanType); ok {
		return c.ColumnTypeScanType(i)
	}
	return reflect.TypeFor[any]()
}

func (r *tracedRows) ColumnTypeNullable(i int) (bool, bool) {
	if c, ok := r.rows.(driver.RowsColumnTypeNullable); ok {
		return c.ColumnTypeNullable(i)
	}
	return false, false
}

func (r *tracedRows) ColumnTypeLength(i int) (int64, bool) {
	if c, ok := r.rows.(driver.RowsColumnTypeLength); ok {
		return c.ColumnTypeLength(i)
	}
	return 0, false
}

func (r *tracedRows) ColumnTypePrecisionScale(i int) (int64, int64, bool) {
	if c, ok := r.rows.(driver.RowsColumnTypePrecisionScale); ok {
		return c.ColumnTypePrecisionScale(i)
	}
	return 0, 0, false
}

type tracedTx struct {
	tx  driver.Tx
	db  string
	ctx context.Context
}

func (t *tracedTx) Commit() error   { return t.run("mysql commit", t.tx.Commit) }
func (t *tracedTx) Rollback() error { return t.run("mysql rollback", t.tx.Rollback) }

func (t *tracedTx) run(op string, f func() error) error {
	span, _ := opentracing.StartSpanFromContext(t.ctx, op)
	err := f()
	end(span, t.db, "", err)
	return err
}

func traceRows(span opentracing.Span, db, query string, rows driver.Rows, err error) (driver.Rows, error) {
	if err != nil {
		end(span, db, query, err)
		return nil, err
	}
	tag(span, db, query, nil)
	return &tracedRows{rows: rows, span: span}, nil
}

func traceExec(span opentracing.Span, db, query string, res driver.Result, err error) (driver.Result, error) {
	if err == nil {
		if n, e := res.RowsAffected(); e == nil {
			span.SetTag("rows_affected", n)
		}
	}
	end(span, db, query, err)
	return res, err
}

func end(span opentracing.Span, db, query string, err error) {
	if err == driver.ErrSkip {
		return
	}
	tag(span, db, query, err)
	span.Finish()
}

func tag(span opentracing.Span, db, query string, err error) {
	span.SetTag("db", db)
	if query != "" {
		span.SetTag("sql", stripLiterals(query))
	}
	fail(span, err)
}

func fail(span opentracing.Span, err error) {
	if err != nil {
		span.SetTag("error", nil)
		span.LogFields(log.Error(err))
	}
}
