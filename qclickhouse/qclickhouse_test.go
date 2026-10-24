package qclickhouse

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	chproto "github.com/ClickHouse/ch-go/proto"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/opentracing/opentracing-go"

	"gopkg.qsoa.cloud/tracer"
)

type hello struct{ db, user, password string }

func fakeProxy(t *testing.T, exception bool) <-chan hello {
	sock := filepath.Join(t.TempDir(), "clickhouse.sock")
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	*sockAddr = "unix://" + sock
	t.Cleanup(func() { *sockAddr = "" })
	hellos := make(chan hello, 10)
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			go serve(c, hellos, exception)
		}
	}()
	return hellos
}

func serve(c net.Conn, hellos chan<- hello, exception bool) {
	defer c.Close()
	r := chproto.NewReader(c)
	if p, err := r.ReadByte(); err != nil || p != 0 {
		return
	}
	r.Str()
	r.UVarInt()
	r.UVarInt()
	r.UVarInt()
	db, _ := r.Str()
	user, _ := r.Str()
	password, err := r.Str()
	if err != nil {
		return
	}
	hellos <- hello{db, user, password}
	var b chproto.Buffer
	if exception {
		b.PutByte(2)
		b.PutInt32(81)
		b.PutString("DB::Exception")
		b.PutString("Database orders does not exist")
		b.PutString("")
		b.PutBool(false)
		c.Write(b.Buf)
		return
	}
	b.PutByte(0)
	b.PutString("fake")
	b.PutUVarInt(24)
	b.PutUVarInt(1)
	b.PutUVarInt(54032)
	c.Write(b.Buf)
	for {
		p, err := r.ReadByte()
		if err != nil || p != 4 {
			return
		}
		c.Write([]byte{4})
	}
}

func TestOpenWithoutSocket(t *testing.T) {
	if _, err := Open(context.Background(), "orders"); !errors.Is(err, ErrNotAttached) || err.Error() != "qclickhouse: no ClickHouse database is attached" {
		t.Fatalf("Open: %v", err)
	}
}

func TestOpenDialsSocket(t *testing.T) {
	hellos := fakeProxy(t, false)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, err := Open(ctx, "orders")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	select {
	case h := <-hellos:
		if h != (hello{db: "orders", user: "default"}) {
			t.Fatalf("hello %+v", h)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("socket was not dialed")
	}
	if tc, ok := c.(*conn); !ok || tc.db != "orders" {
		t.Fatalf("Open returned %T", c)
	}
}

func goroutines() int {
	for n := runtime.NumGoroutine(); ; {
		time.Sleep(20 * time.Millisecond)
		m := runtime.NumGoroutine()
		if m == n {
			return n
		}
		n = m
	}
}

func TestOpenUnknownDatabase(t *testing.T) {
	fakeProxy(t, true)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	before := goroutines()
	_, err := Open(ctx, "orders")
	var e *clickhouse.Exception
	if !errors.As(err, &e) || e.Code != 81 {
		t.Fatalf("Open: %v", err)
	}
	for deadline := time.Now().Add(time.Second); runtime.NumGoroutine() > before; time.Sleep(10 * time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatal("the pool of a failed Open was not closed")
		}
	}
}

var (
	errUnknownTable = &clickhouse.Exception{Code: 60, Name: "DB::Exception", Message: "Table orders.hunter2 does not exist"}
	errBoom         = fmt.Errorf("cannot parse 'hunter2': %w", errors.New("boom"))
)

func result(q string) error {
	switch q {
	case "unknown":
		return errUnknownTable
	case "boom":
		return errBoom
	case "none":
		return sql.ErrNoRows
	}
	return nil
}

type fakeConn struct {
	driver.Conn
	seen *[]context.Context
}

func (f fakeConn) saw(ctx context.Context) {
	if f.seen != nil {
		*f.seen = append(*f.seen, ctx)
	}
}

func (f fakeConn) Query(ctx context.Context, q string, _ ...any) (driver.Rows, error) {
	f.saw(ctx)
	if q == "broken" {
		return &fakeRows{err: errBoom}, nil
	}
	if err := result(q); err != nil {
		return nil, err
	}
	return &fakeRows{}, nil
}

func (f fakeConn) QueryRow(ctx context.Context, q string, _ ...any) driver.Row {
	f.saw(ctx)
	if q == "unknown" {
		return &fakeRow{err: errUnknownTable}
	}
	return &fakeRow{scan: result(q)}
}

func (f fakeConn) Select(ctx context.Context, _ any, q string, _ ...any) error {
	f.saw(ctx)
	return result(q)
}

func (f fakeConn) Exec(ctx context.Context, q string, _ ...any) error {
	f.saw(ctx)
	return result(q)
}

func (f fakeConn) AsyncInsert(ctx context.Context, q string, _ bool, _ ...any) error {
	f.saw(ctx)
	return result(q)
}

func (f fakeConn) PrepareBatch(ctx context.Context, q string, _ ...driver.PrepareBatchOption) (driver.Batch, error) {
	f.saw(ctx)
	if err := result(q); err != nil {
		return nil, err
	}
	return &fakeBatch{}, nil
}

type fakeRows struct {
	driver.Rows
	n   int
	err error
}

func (r *fakeRows) Next() bool {
	if r.n == 3 {
		return false
	}
	r.n++
	return true
}
func (r *fakeRows) Err() error   { return r.err }
func (r *fakeRows) Close() error { return nil }

type fakeRow struct {
	driver.Row
	err, scan error
}

func (r *fakeRow) Err() error                { return r.err }
func (r *fakeRow) Scan(...any) error         { return r.scan }
func (r *fakeRow) ScanStruct(dest any) error { return r.scan }

type fakeBatch struct {
	driver.Batch
	n int
}

func (b *fakeBatch) Append(...any) error { b.n++; return nil }
func (b *fakeBatch) Rows() int           { return b.n }
func (b *fakeBatch) Flush() error        { b.n = 0; return nil }
func (b *fakeBatch) Send() error         { b.n = 0; return nil }
func (b *fakeBatch) Abort() error        { return nil }
func (b *fakeBatch) Close() error        { return nil }

type span struct {
	O string `json:"o"`
	C struct {
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

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func ops(got []span) []string {
	var out []string
	for _, s := range got {
		out = append(out, s.O)
	}
	return out
}

func TestQuerySpan(t *testing.T) {
	buf := traced(t)
	ctx, pid := parent()
	c := &conn{Conn: fakeConn{}, db: "orders"}
	rows := must(c.Query(ctx, "SELECT a FROM t WHERE p = 'hunter2'"))
	for rows.Next() {
	}
	rows.Close()
	if err := c.QueryRow(ctx, "SELECT a FROM t WHERE id = 7").Scan(); err != nil {
		t.Fatal(err)
	}
	must(0, c.Select(ctx, nil, "SELECT a FROM t WHERE p = 'hunter2'"))
	must(0, c.Exec(ctx, "ALTER TABLE t DELETE WHERE p = 'hunter2'"))
	must(0, c.AsyncInsert(ctx, "INSERT INTO t VALUES ('hunter2')", false))
	b := must(c.PrepareBatch(ctx, "INSERT INTO t"))
	b.Append(1)
	must(0, b.Flush())
	b.Append(2)
	b.Append(3)
	must(0, b.Send())
	b.Close()
	if strings.Contains(buf.String(), "hunter2") {
		t.Fatalf("literal in span: %s", buf)
	}
	got := spans(t, buf)
	want := []map[string]any{
		{"db": "orders", "sql": "SELECT a FROM t WHERE p = ?", "rows": 3.0},
		{"db": "orders", "sql": "SELECT a FROM t WHERE id = ?"},
		{"db": "orders", "sql": "SELECT a FROM t WHERE p = ?"},
		{"db": "orders", "sql": "ALTER TABLE t DELETE WHERE p = ?"},
		{"db": "orders", "sql": "INSERT INTO t VALUES (?)"},
		{"db": "orders", "sql": "INSERT INTO t", "rows": 3.0},
	}
	wantOps := []string{"clickhouse query", "clickhouse query row", "clickhouse select", "clickhouse exec", "clickhouse async insert", "clickhouse batch"}
	if !slices.Equal(ops(got), wantOps) {
		t.Fatalf("spans %v, want %v", ops(got), wantOps)
	}
	for i, s := range got {
		if s.C.P != pid || !reflect.DeepEqual(s.T, want[i]) {
			t.Errorf("%s: parent %d tags %v, want parent %d tags %v", s.O, s.C.P, s.T, pid, want[i])
		}
	}
}

func TestSpanEnds(t *testing.T) {
	buf := traced(t)
	ctx := context.Background()
	c := &conn{Conn: fakeConn{}, db: "orders"}
	must(c.Query(ctx, "SELECT 1")).Close()
	for rows := must(c.Query(ctx, "SELECT 2")); rows.Next(); {
	}
	row := c.QueryRow(ctx, "none")
	if err := row.Scan(); err != sql.ErrNoRows {
		t.Fatalf("scan: %v", err)
	}
	row.Scan()
	c.QueryRow(ctx, "SELECT 3").ScanStruct(nil)
	must(c.PrepareBatch(ctx, "INSERT INTO a")).Abort()
	must(c.PrepareBatch(ctx, "INSERT INTO b")).Close()
	got := spans(t, buf)
	wantOps := []string{"clickhouse query", "clickhouse query", "clickhouse query row", "clickhouse query row", "clickhouse batch", "clickhouse batch"}
	if !slices.Equal(ops(got), wantOps) {
		t.Fatalf("spans %v, want %v", ops(got), wantOps)
	}
	for _, s := range got {
		if _, ok := s.T["error"]; ok {
			t.Errorf("%s: error tag %v", s.O, s.T)
		}
	}
	if got[0].T["rows"] != 0.0 || got[1].T["rows"] != 3.0 {
		t.Errorf("rows %v, %v", got[0].T["rows"], got[1].T["rows"])
	}
}

func TestSpanErrors(t *testing.T) {
	buf := traced(t)
	ctx := context.Background()
	c := &conn{Conn: fakeConn{}, db: "orders"}
	if _, err := c.Query(ctx, "unknown"); err != errUnknownTable {
		t.Fatalf("query: %v", err)
	}
	for rows := must(c.Query(ctx, "broken")); rows.Next(); {
	}
	if err := c.QueryRow(ctx, "unknown").Err(); err != errUnknownTable {
		t.Fatalf("query row: %v", err)
	}
	if err := c.QueryRow(ctx, "boom").Scan(); err != errBoom {
		t.Fatalf("scan: %v", err)
	}
	if err := c.Select(ctx, nil, "boom"); err != errBoom {
		t.Fatalf("select: %v", err)
	}
	if err := c.Exec(ctx, "unknown"); err != errUnknownTable {
		t.Fatalf("exec: %v", err)
	}
	if err := c.AsyncInsert(ctx, "boom", true); err != errBoom {
		t.Fatalf("async insert: %v", err)
	}
	if _, err := c.PrepareBatch(ctx, "unknown"); err != errUnknownTable {
		t.Fatalf("batch: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "hunter2") || strings.Contains(out, "does not exist") {
		t.Errorf("data in span: %s", out)
	}
	for _, want := range []string{`"v":"Code 60 (UNKNOWN_TABLE)"`, `"v":"cannot parse ?: boom"`} {
		if !strings.Contains(out, want) {
			t.Errorf("no %s in %s", want, out)
		}
	}
	got := spans(t, bytes.NewBufferString(out))
	if len(got) != 8 {
		t.Fatalf("spans %v", ops(got))
	}
	for _, s := range got {
		if _, ok := s.T["error"]; !ok {
			t.Errorf("%s: no error tag: %v", s.O, s.T)
		}
	}
}

func TestHeredocNotTraced(t *testing.T) {
	buf := traced(t)
	ctx := context.Background()
	c := &conn{Conn: fakeConn{}, db: "orders"}
	must(c.Query(ctx, "SELECT $$hunter2$$")).Close()
	must(0, c.Exec(ctx, "INSERT INTO t VALUES ($x$hunter2$x$)"))
	if len(spans(t, bytes.NewBufferString(buf.String()))) != 2 || strings.Contains(buf.String(), "hunter2") {
		t.Fatalf("spans: %s", buf)
	}
}

func TestCommentNotTraced(t *testing.T) {
	buf := traced(t)
	c := &conn{Conn: fakeConn{}, db: "orders"}
	must(0, c.Exec(context.Background(), "SELECT a FROM t --don't use the index\nWHERE token = 'hunter2'"))
	if len(spans(t, bytes.NewBufferString(buf.String()))) != 1 || strings.Contains(buf.String(), "hunter2") {
		t.Fatalf("spans: %s", buf)
	}
}

func TestDriverContext(t *testing.T) {
	var seen []context.Context
	c := &conn{Conn: fakeConn{seen: &seen}, db: "orders"}
	type key struct{}
	ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), key{}, "v"), time.Minute)
	defer cancel()
	must(c.Query(ctx, "SELECT 1")).Close()
	c.QueryRow(ctx, "SELECT 1").Scan()
	must(0, c.Select(ctx, nil, "SELECT 1"))
	must(0, c.Exec(ctx, "SELECT 1"))
	must(0, c.AsyncInsert(ctx, "INSERT INTO t VALUES (1)", false))
	must(0, must(c.PrepareBatch(ctx, "INSERT INTO t")).Send())
	if len(seen) != 6 {
		t.Fatalf("%d driver calls", len(seen))
	}
	cancel()
	for i, d := range seen {
		if _, ok := d.Deadline(); ok || d.Done() != ctx.Done() || d.Err() != context.Canceled || d.Value(key{}) != "v" {
			t.Errorf("call %d: the driver context differs from the caller's beyond its deadline", i)
		}
	}
	seen = nil
	must(c.Query(context.Background(), "SELECT 1")).Close()
	if _, ok := seen[0].(noDeadline); ok {
		t.Error("a context without a deadline was wrapped")
	}
}

func clickhouseTest(t *testing.T) (driver.Conn, *clickhouse.Options) {
	dsn := os.Getenv("QSOA_TEST_CLICKHOUSE_DSN")
	if dsn == "" {
		t.Skip("QSOA_TEST_CLICKHOUSE_DSN is not set")
	}
	opts, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		t.Fatal("cannot parse QSOA_TEST_CLICKHOUSE_DSN")
	}
	base := must(clickhouse.Open(opts))
	t.Cleanup(func() { base.Close() })
	return base, opts
}

func TestDeadlineNotSent(t *testing.T) {
	base, opts := clickhouseTest(t)
	name := fmt.Sprintf("qclickhouse_test_%d", time.Now().UnixNano())
	password := rand.Text()
	t.Cleanup(func() {
		base.Exec(context.Background(), "DROP USER IF EXISTS "+name)
		base.Exec(context.Background(), "DROP SETTINGS PROFILE IF EXISTS "+name)
		base.Exec(context.Background(), "DROP DATABASE IF EXISTS "+name)
	})
	for _, q := range []string{
		"CREATE SETTINGS PROFILE " + name + " SETTINGS max_execution_time = 60 CONST",
		"CREATE USER " + name + " IDENTIFIED WITH sha256_password BY '" + password + "' HOST ANY SETTINGS PROFILE '" + name + "'",
		"CREATE DATABASE " + name,
		"CREATE TABLE " + name + ".t (id UInt64) ENGINE = MergeTree ORDER BY id",
		"GRANT SELECT, INSERT ON " + name + ".* TO " + name,
	} {
		if err := base.Exec(context.Background(), q); err != nil {
			t.Fatalf("setup: %s", errText(err))
		}
	}
	uo := *opts
	uo.Auth = clickhouse.Auth{Username: name, Password: password}
	user := must(clickhouse.Open(&uo))
	t.Cleanup(func() { user.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var e *clickhouse.Exception
	if err := user.Exec(ctx, "SELECT 1"); !errors.As(err, &e) || e.Code != 452 {
		t.Fatalf("the driver alone: %v", err)
	}
	c := &conn{Conn: user, db: "default"}
	var n uint8
	if err := c.QueryRow(ctx, "SELECT 1").Scan(&n); err != nil || n != 1 {
		t.Fatalf("query with a deadline: %d, %v", n, err)
	}
	if err := c.Exec(ctx, "INSERT INTO "+name+".t VALUES (1)"); err != nil {
		t.Fatalf("exec with a deadline: %v", err)
	}
	b, err := c.PrepareBatch(ctx, "INSERT INTO "+name+".t")
	if err != nil {
		t.Fatalf("batch with a deadline: %v", err)
	}
	must(0, b.Append(uint64(2)))
	if err := b.Send(); err != nil {
		t.Fatalf("batch send with a deadline: %v", err)
	}
	var rows uint64
	if err := c.QueryRow(ctx, "SELECT count() FROM "+name+".t").Scan(&rows); err != nil || rows != 2 {
		t.Fatalf("rows %d, %v", rows, err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	start := time.Now()
	var count uint64
	err = c.QueryRow(ctx, "SELECT count() FROM numbers(1000000000000)").Scan(&count)
	if d := time.Since(start); !errors.Is(err, context.DeadlineExceeded) || d > 10*time.Second {
		t.Fatalf("long query: %v after %s", err, d)
	}
}

func TestClickHouseSpans(t *testing.T) {
	base, _ := clickhouseTest(t)
	ctx := context.Background()
	db := fmt.Sprintf("qclickhouse_test_%d", time.Now().UnixNano())
	if err := base.Exec(ctx, "CREATE DATABASE "+db); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { base.Exec(context.Background(), "DROP DATABASE IF EXISTS "+db) })
	buf := traced(t)
	c := &conn{Conn: base, db: db}
	must(0, c.Exec(ctx, "CREATE TABLE "+db+".t (id UInt64, p String) ENGINE = MergeTree ORDER BY id"))
	b := must(c.PrepareBatch(ctx, "INSERT INTO "+db+".t"))
	must(0, b.Append(uint64(1), "hunter2"))
	must(0, b.Flush())
	must(0, b.Append(uint64(2), "x"))
	must(0, b.Append(uint64(3), "y"))
	must(0, b.Send())
	rows := must(c.Query(ctx, "SELECT id FROM "+db+".t WHERE p != 'hunter2' ORDER BY id"))
	for rows.Next() {
	}
	rows.Close()
	var n uint64
	if err := c.QueryRow(ctx, "SELECT count() FROM "+db+".t").Scan(&n); err != nil || n != 3 {
		t.Fatalf("count %d, %v", n, err)
	}
	if _, err := c.Query(ctx, "SELECT 'hunter2' FROM "+db+".missing"); err == nil {
		t.Fatal("query of a missing table succeeded")
	}
	out := buf.String()
	if strings.Contains(out, "hunter2") || !strings.Contains(out, `"v":"Code 60 (UNKNOWN_TABLE)"`) {
		t.Errorf("spans: %s", out)
	}
	got := spans(t, bytes.NewBufferString(out))
	wantOps := []string{"clickhouse exec", "clickhouse batch", "clickhouse query", "clickhouse query row", "clickhouse query"}
	if !slices.Equal(ops(got), wantOps) {
		t.Fatalf("spans %v, want %v", ops(got), wantOps)
	}
	if got[1].T["rows"] != 3.0 || got[2].T["rows"] != 2.0 {
		t.Errorf("rows %v, %v", got[1].T["rows"], got[2].T["rows"])
	}
}

const benchSQL = "SELECT id, name FROM events WHERE status = 'active' AND ts > '2026-01-01 00:00:00' AND score >= 42.5 ORDER BY ts DESC LIMIT 100"

func tracedQuery(ctx context.Context) func() {
	c := &conn{Conn: fakeConn{}, db: "orders"}
	return func() {
		rows := must(c.Query(ctx, benchSQL))
		for rows.Next() {
		}
		rows.Close()
	}
}

func BenchmarkTracedQuery(b *testing.B) {
	q := tracedQuery(context.Background())
	b.ReportAllocs()
	for b.Loop() {
		q()
	}
}

func BenchmarkTracedQueryParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		q := tracedQuery(context.Background())
		for pb.Next() {
			q()
		}
	})
}

func BenchmarkTracedQueryDeadline(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	q := tracedQuery(ctx)
	b.ReportAllocs()
	for b.Loop() {
		q()
	}
}

func BenchmarkTracedQueryDeadlineParallel(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		q := tracedQuery(ctx)
		for pb.Next() {
			q()
		}
	})
}

func sendBatch() func() {
	c := &conn{Conn: fakeConn{}, db: "orders"}
	ctx := context.Background()
	return func() {
		b := must(c.PrepareBatch(ctx, "INSERT INTO events"))
		b.Append(1)
		must(0, b.Send())
	}
}

func BenchmarkTracedBatch(b *testing.B) {
	s := sendBatch()
	b.ReportAllocs()
	for b.Loop() {
		s()
	}
}

func BenchmarkTracedBatchParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		s := sendBatch()
		for pb.Next() {
			s()
		}
	})
}
