package qmysql

import (
	"context"
	"database/sql/driver"
	"testing"
)

const benchSQL = "SELECT id, name, email FROM users WHERE status = 'active' AND created_at > '2026-01-01 00:00:00' " +
	"AND score >= 42.5 AND `group` IN (1, 2, 3) AND token = x'0a0b' ORDER BY created_at DESC LIMIT 100 OFFSET 20"

func BenchmarkStripLiterals(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		stripLiterals(benchSQL)
	}
}

func BenchmarkStripLiteralsParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			stripLiterals(benchSQL)
		}
	})
}

func tracedQuery() func() {
	q := connect(nil).(driver.QueryerContext)
	dest := make([]driver.Value, 1)
	ctx := context.Background()
	return func() {
		rows := must(q.QueryContext(ctx, benchSQL, nil))
		for rows.Next(dest) == nil {
		}
		rows.Close()
	}
}

func BenchmarkTracedQuery(b *testing.B) {
	q := tracedQuery()
	b.ReportAllocs()
	for b.Loop() {
		q()
	}
}

func BenchmarkTracedQueryParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		q := tracedQuery()
		for pb.Next() {
			q()
		}
	})
}
