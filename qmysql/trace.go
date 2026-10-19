package qmysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"reflect"
	"strconv"

	"github.com/go-sql-driver/mysql"
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
	span, sctx := opentracing.StartSpanFromContext(ctx, "mysql begin")
	tx, err := c.begin(sctx, opts)
	end(span, c.db, "", err)
	if err != nil {
		return nil, err
	}
	return &tracedTx{tx: tx, db: c.db, ctx: ctx}, nil
}

func (c *tracedConn) begin(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if b, ok := c.conn.(driver.ConnBeginTx); ok {
		return b.BeginTx(ctx, opts)
	}
	if opts.Isolation != driver.IsolationLevel(sql.LevelDefault) {
		return nil, errors.New("sql: driver does not support non-default isolation level")
	}
	if opts.ReadOnly {
		return nil, errors.New("sql: driver does not support read-only transactions")
	}
	tx, err := c.conn.Begin()
	if err == nil && ctx.Err() != nil {
		tx.Rollback()
		return nil, ctx.Err()
	}
	return tx, err
}

func (c *tracedConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "mysql prepare")
	stmt, err := c.prepare(ctx, query)
	stripped := stripLiterals(query)
	end(span, c.db, stripped, err)
	if err != nil {
		return nil, err
	}
	return &tracedStmt{stmt: stmt, db: c.db, sql: stripped}, nil
}

func (c *tracedConn) prepare(ctx context.Context, query string) (driver.Stmt, error) {
	if p, ok := c.conn.(driver.ConnPrepareContext); ok {
		return p.PrepareContext(ctx, query)
	}
	stmt, err := c.conn.Prepare(query)
	if err == nil && ctx.Err() != nil {
		stmt.Close()
		return nil, ctx.Err()
	}
	return stmt, err
}

func (c *tracedConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	q, ok := c.conn.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	span, ctx := opentracing.StartSpanFromContext(ctx, "mysql query")
	rows, err := q.QueryContext(ctx, query, args)
	if err == driver.ErrSkip {
		return nil, err
	}
	return traceRows(span, c.db, stripLiterals(query), rows, err)
}

func (c *tracedConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	e, ok := c.conn.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	span, ctx := opentracing.StartSpanFromContext(ctx, "mysql exec")
	res, err := e.ExecContext(ctx, query, args)
	if err == driver.ErrSkip {
		return nil, err
	}
	return traceExec(span, c.db, stripLiterals(query), res, err)
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
	stmt driver.Stmt
	db   string
	sql  string
}

func (s *tracedStmt) Close() error  { return s.stmt.Close() }
func (s *tracedStmt) NumInput() int { return s.stmt.NumInput() }

func (s *tracedStmt) Exec(args []driver.Value) (driver.Result, error) { return s.stmt.Exec(args) }
func (s *tracedStmt) Query(args []driver.Value) (driver.Rows, error)  { return s.stmt.Query(args) }

func (s *tracedStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "mysql stmt query")
	rows, err := s.rows(ctx, args)
	return traceRows(span, s.db, s.sql, rows, err)
}

func (s *tracedStmt) rows(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	if q, ok := s.stmt.(driver.StmtQueryContext); ok {
		return q.QueryContext(ctx, args)
	}
	v, err := values(ctx, args)
	if err != nil {
		return nil, err
	}
	return s.stmt.Query(v)
}

func (s *tracedStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "mysql stmt exec")
	res, err := s.exec(ctx, args)
	return traceExec(span, s.db, s.sql, res, err)
}

func (s *tracedStmt) exec(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	if e, ok := s.stmt.(driver.StmtExecContext); ok {
		return e.ExecContext(ctx, args)
	}
	v, err := values(ctx, args)
	if err != nil {
		return nil, err
	}
	return s.stmt.Exec(v)
}

func values(ctx context.Context, named []driver.NamedValue) ([]driver.Value, error) {
	v := make([]driver.Value, len(named))
	for i, p := range named {
		if p.Name != "" {
			return nil, errors.New("sql: driver does not support the use of Named Parameters")
		}
		v[i] = p.Value
	}
	return v, ctx.Err()
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

func traceRows(span opentracing.Span, db, stripped string, rows driver.Rows, err error) (driver.Rows, error) {
	if err != nil {
		end(span, db, stripped, err)
		return nil, err
	}
	tag(span, db, stripped, nil)
	return &tracedRows{rows: rows, span: span}, nil
}

func traceExec(span opentracing.Span, db, stripped string, res driver.Result, err error) (driver.Result, error) {
	if err == nil {
		if n, e := res.RowsAffected(); e == nil {
			span.SetTag("rows_affected", n)
		}
	}
	end(span, db, stripped, err)
	return res, err
}

func end(span opentracing.Span, db, stripped string, err error) {
	if err == driver.ErrSkip {
		return
	}
	tag(span, db, stripped, err)
	span.Finish()
}

func tag(span opentracing.Span, db, stripped string, err error) {
	span.SetTag("db", db)
	if stripped != "" {
		span.SetTag("sql", stripped)
	}
	fail(span, err)
}

func fail(span opentracing.Span, err error) {
	if err != nil {
		span.SetTag("error", nil)
		span.LogFields(log.String("error.object", errText(err)))
	}
}

func errText(err error) string {
	var me *mysql.MySQLError
	if !errors.As(err, &me) {
		return stripLiterals(err.Error())
	}
	s := "Error " + strconv.Itoa(int(me.Number))
	if me.SQLState != [5]byte{} {
		s += " (" + string(me.SQLState[:]) + ")"
	}
	return s
}
