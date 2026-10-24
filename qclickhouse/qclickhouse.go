// Package qclickhouse opens the ClickHouse databases attached to the service.
package qclickhouse

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

var sockAddr = flag.String("q_clickhouse_sock", "", "ClickHouse socket")

var errNoSocket = errors.New("qclickhouse: no ClickHouse database is attached")

// Open connects to the attached database name through the runner and pings it. The Conn traces every query.
func Open(ctx context.Context, name string) (driver.Conn, error) {
	if !flag.Parsed() {
		flag.Parse()
	}
	if *sockAddr == "" {
		return nil, errNoSocket
	}
	path := strings.TrimPrefix(*sockAddr, "unix://")
	c, err := clickhouse.Open(&clickhouse.Options{
		Auth: clickhouse.Auth{Database: name},
		DialContext: func(ctx context.Context, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", path)
		},
	})
	if err != nil {
		return nil, err
	}
	if err := c.Ping(ctx); err != nil {
		c.Close()
		return nil, fmt.Errorf("qclickhouse: %s: %w", name, err)
	}
	return &conn{Conn: c, db: name}, nil
}
