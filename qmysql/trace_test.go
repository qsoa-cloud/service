package qmysql

import (
	"bytes"
	"cmp"
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/opentracing/opentracing-go"

	"gopkg.qsoa.cloud/tracer"
)

type calls struct {
	mu sync.Mutex
	m  map[string]int
}

func (c *calls) add(name string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.m == nil {
		c.m = map[string]int{}
	}
	c.m[name]++
}

func (c *calls) saw(name string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m[name] > 0
}

var errBoom = errors.New("boom")

func result(query string) error {
	switch query {
	case "skip":
		return driver.ErrSkip
	case "fail":
		return errBoom
	}
	return nil
}

type fakeConnector struct{ c *calls }

func (f fakeConnector) Connect(context.Context) (driver.Conn, error) { return &fakeConn{f.c}, nil }
func (fakeConnector) Driver() driver.Driver                          { return nil }

type fakeConn struct{ c *calls }

func (f *fakeConn) Prepare(q string) (driver.Stmt, error) {
	return f.PrepareContext(context.Background(), q)
}
func (f *fakeConn) Close() error { return nil }
func (f *fakeConn) Begin() (driver.Tx, error) {
	return f.BeginTx(context.Background(), driver.TxOptions{})
}
func (f *fakeConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return fakeTx{}, nil
}
func (f *fakeConn) PrepareContext(_ context.Context, q string) (driver.Stmt, error) {
	if err := result(q); err != nil {
		return nil, err
	}
	return &fakeStmt{f.c}, nil
}
func (f *fakeConn) QueryContext(_ context.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
	if err := result(q); err != nil {
		return nil, err
	}
	if q == "broken" {
		return &fakeRows{c: f.c, err: errBoom}, nil
	}
	return &fakeRows{c: f.c}, nil
}
func (f *fakeConn) ExecContext(_ context.Context, q string, _ []driver.NamedValue) (driver.Result, error) {
	if err := result(q); err != nil {
		return nil, err
	}
	return driver.RowsAffected(5), nil
}
func (f *fakeConn) Ping(context.Context) error         { f.c.add("ping"); return nil }
func (f *fakeConn) ResetSession(context.Context) error { f.c.add("reset"); return nil }
func (f *fakeConn) IsValid() bool                      { f.c.add("valid"); return true }
func (f *fakeConn) CheckNamedValue(*driver.NamedValue) error {
	f.c.add("check")
	return nil
}

type fakeStmt struct{ c *calls }

func (s *fakeStmt) Close() error  { return nil }
func (s *fakeStmt) NumInput() int { return -1 }
func (s *fakeStmt) Exec([]driver.Value) (driver.Result, error) {
	return driver.RowsAffected(5), nil
}
func (s *fakeStmt) Query([]driver.Value) (driver.Rows, error) { return &fakeRows{c: s.c}, nil }
func (s *fakeStmt) QueryContext(context.Context, []driver.NamedValue) (driver.Rows, error) {
	return &fakeRows{c: s.c}, nil
}
func (s *fakeStmt) ExecContext(context.Context, []driver.NamedValue) (driver.Result, error) {
	return driver.RowsAffected(5), nil
}
func (s *fakeStmt) CheckNamedValue(*driver.NamedValue) error {
	s.c.add("stmt check")
	return nil
}
func (s *fakeStmt) ColumnConverter(int) driver.ValueConverter {
	s.c.add("converter")
	return driver.DefaultParameterConverter
}

type fakeRows struct {
	c   *calls
	n   int
	err error
}

func (r *fakeRows) Columns() []string { return []string{"a"} }
func (r *fakeRows) Close() error      { return nil }
func (r *fakeRows) Next(dest []driver.Value) error {
	if r.n == 3 {
		return cmp.Or(r.err, io.EOF)
	}
	r.n++
	dest[0] = int64(r.n)
	return nil
}
func (r *fakeRows) HasNextResultSet() bool { r.c.add("has next"); return false }
func (r *fakeRows) NextResultSet() error   { r.c.add("next set"); return io.EOF }
func (r *fakeRows) ColumnTypeDatabaseTypeName(int) string {
	r.c.add("type name")
	return "INT"
}
func (r *fakeRows) ColumnTypeScanType(int) reflect.Type {
	r.c.add("scan type")
	return reflect.TypeFor[int64]()
}
func (r *fakeRows) ColumnTypeNullable(int) (bool, bool) { r.c.add("nullable"); return true, true }
func (r *fakeRows) ColumnTypeLength(int) (int64, bool)  { r.c.add("length"); return 4, true }
func (r *fakeRows) ColumnTypePrecisionScale(int) (int64, int64, bool) {
	r.c.add("precision")
	return 1, 2, true
}

type fakeTx struct{}

func (fakeTx) Commit() error   { return nil }
func (fakeTx) Rollback() error { return nil }

type span struct {
	O string `json:"o"`
	C struct {
		S uint64 `json:"s"`
		P uint64 `json:"p"`
	} `json:"c"`
	T map[string]any `json:"t"`
}

func traced(t *testing.T) *bytes.Buffer {
	buf := &bytes.Buffer{}
	opentracing.SetGlobalTracer(tracer.New(buf))
	t.Cleanup(func() { opentracing.SetGlobalTracer(opentracing.NoopTracer{}) })
	return buf
}

func spans(t *testing.T, buf *bytes.Buffer) []span {
	var out []span
	for d := json.NewDecoder(buf); d.More(); {
		var s span
		if err := d.Decode(&s); err != nil {
			t.Fatal(err)
		}
		out = append(out, s)
	}
	return out
}

func parent() (context.Context, uint64) {
	p := opentracing.StartSpan("parent")
	return opentracing.ContextWithSpan(context.Background(), p), p.Context().(*tracer.SpanContext).SpanID
}

func connect(c *calls) driver.Conn {
	return must((&tracedConnector{base: fakeConnector{c}, db: "orders"}).Connect(context.Background()))
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func TestTraceQueryAndExec(t *testing.T) {
	buf := traced(t)
	ctx, pid := parent()
	conn := connect(nil)
	rows := must(conn.(driver.QueryerContext).QueryContext(ctx, "SELECT a FROM users WHERE password = 'hunter2'", nil))
	for dest := make([]driver.Value, 1); rows.Next(dest) == nil; {
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	must(conn.(driver.ExecerContext).ExecContext(ctx, "UPDATE users SET password = 'hunter2' WHERE id = 7", nil))
	if strings.Contains(buf.String(), "hunter2") {
		t.Fatalf("literal in span: %s", buf)
	}
	got := spans(t, buf)
	want := []map[string]any{
		{"db": "orders", "sql": "SELECT a FROM users WHERE password = ?", "rows": 3.0},
		{"db": "orders", "sql": "UPDATE users SET password = ? WHERE id = ?", "rows_affected": 5.0},
	}
	if len(got) != 2 || got[0].O != "mysql query" || got[1].O != "mysql exec" {
		t.Fatalf("spans %+v", got)
	}
	for i, s := range got {
		if s.C.P != pid || !reflect.DeepEqual(s.T, want[i]) {
			t.Errorf("%s: parent %d tags %v, want parent %d tags %v", s.O, s.C.P, s.T, pid, want[i])
		}
	}
}

func TestTraceStmtAndTx(t *testing.T) {
	buf := traced(t)
	ctx, pid := parent()
	conn := connect(nil)
	stmt := must(conn.(driver.ConnPrepareContext).PrepareContext(ctx, "SELECT a FROM t WHERE b = 'x' AND c = ?"))
	if err := must(stmt.(driver.StmtQueryContext).QueryContext(ctx, nil)).Close(); err != nil {
		t.Fatal(err)
	}
	must(stmt.(driver.StmtExecContext).ExecContext(ctx, nil))
	if err := must(conn.(driver.ConnBeginTx).BeginTx(ctx, driver.TxOptions{})).Commit(); err != nil {
		t.Fatal(err)
	}
	if err := must(conn.(driver.ConnBeginTx).BeginTx(ctx, driver.TxOptions{})).Rollback(); err != nil {
		t.Fatal(err)
	}
	var ops []string
	for _, s := range spans(t, buf) {
		ops = append(ops, s.O)
		if s.C.P != pid || s.T["db"] != "orders" {
			t.Errorf("%s: parent %d tags %v", s.O, s.C.P, s.T)
		}
		if sql, ok := s.T["sql"]; ok != (s.O == "mysql prepare" || strings.HasPrefix(s.O, "mysql stmt")) || ok && sql != "SELECT a FROM t WHERE b = ? AND c = ?" {
			t.Errorf("%s: sql %v", s.O, sql)
		}
	}
	want := []string{"mysql prepare", "mysql stmt query", "mysql stmt exec", "mysql begin", "mysql commit", "mysql begin", "mysql rollback"}
	if !slices.Equal(ops, want) {
		t.Fatalf("spans %v, want %v", ops, want)
	}
}

func TestTraceSkip(t *testing.T) {
	buf := traced(t)
	conn := connect(nil)
	ctx := context.Background()
	if _, err := conn.(driver.QueryerContext).QueryContext(ctx, "skip", nil); err != driver.ErrSkip {
		t.Fatalf("query: %v", err)
	}
	if _, err := conn.(driver.ExecerContext).ExecContext(ctx, "skip", nil); err != driver.ErrSkip {
		t.Fatalf("exec: %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("spans for ErrSkip: %s", buf)
	}
}

func TestTraceError(t *testing.T) {
	buf := traced(t)
	q := connect(nil).(driver.QueryerContext)
	if _, err := q.QueryContext(context.Background(), "fail", nil); err != errBoom {
		t.Fatalf("query: %v", err)
	}
	rows := must(q.QueryContext(context.Background(), "broken", nil))
	for dest := make([]driver.Value, 1); rows.Next(dest) == nil; {
	}
	rows.Close()
	got := spans(t, buf)
	if len(got) != 2 {
		t.Fatalf("spans %+v", got)
	}
	for _, s := range got {
		if _, ok := s.T["error"]; s.O != "mysql query" || !ok {
			t.Errorf("%s: no error tag: %v", s.O, s.T)
		}
	}
}

func TestForwarding(t *testing.T) {
	c := &calls{}
	conn := connect(c)
	ctx := context.Background()
	conn.(driver.NamedValueChecker).CheckNamedValue(&driver.NamedValue{})
	conn.(driver.SessionResetter).ResetSession(ctx)
	conn.(driver.Validator).IsValid()
	conn.(driver.Pinger).Ping(ctx)
	stmt := must(conn.(driver.ConnPrepareContext).PrepareContext(ctx, "SELECT 1"))
	stmt.(driver.NamedValueChecker).CheckNamedValue(&driver.NamedValue{})
	stmt.(driver.ColumnConverter).ColumnConverter(0)
	rows := must(conn.(driver.QueryerContext).QueryContext(ctx, "SELECT 1", nil))
	rows.(driver.RowsNextResultSet).HasNextResultSet()
	rows.(driver.RowsNextResultSet).NextResultSet()
	rows.(driver.RowsColumnTypeDatabaseTypeName).ColumnTypeDatabaseTypeName(0)
	rows.(driver.RowsColumnTypeScanType).ColumnTypeScanType(0)
	rows.(driver.RowsColumnTypeNullable).ColumnTypeNullable(0)
	rows.(driver.RowsColumnTypeLength).ColumnTypeLength(0)
	rows.(driver.RowsColumnTypePrecisionScale).ColumnTypePrecisionScale(0)
	rows.Close()
	for _, n := range []string{"check", "reset", "valid", "ping", "stmt check", "converter", "has next", "next set", "type name", "scan type", "nullable", "length", "precision"} {
		if !c.saw(n) {
			t.Errorf("%s was not forwarded", n)
		}
	}
}
