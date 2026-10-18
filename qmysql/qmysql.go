//go:generate protoc -I internal/pb --gogofaster_out=plugins=grpc:internal/pb internal/pb/qmysql.proto
package qmysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"flag"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"

	"gopkg.qsoa.cloud/service/internal/qconn"
	"gopkg.qsoa.cloud/service/qmysql/internal/pb"
)

var (
	addr     = flag.String("q_mysql_addr", "", "mysql discovery address")
	sockAddr = flag.String("q_mysql_sock", "", "mysql proxy socket")
)

type qDriver struct {
	mtx    sync.Mutex
	client pb.MySqlClient
}

func init() {
	sql.Register("qmysql", &qDriver{})
}

func (d *qDriver) Open(name string) (driver.Conn, error) {
	c, err := d.OpenConnector(name)
	if err != nil {
		return nil, err
	}
	return c.Connect(context.Background())
}

func (d *qDriver) OpenConnector(name string) (driver.Connector, error) {
	if !flag.Parsed() {
		flag.Parse()
	}
	if *sockAddr != "" {
		c, err := mysql.NewConnector(proxyConfig(*sockAddr, name))
		if err != nil {
			return nil, err
		}
		return &tracedConnector{base: c, drv: d, db: name}, nil
	}
	return &tracedConnector{base: &dsnConnector{d: d, name: name}, drv: d, db: name}, nil
}

func proxyConfig(sock, name string) *mysql.Config {
	cfg := mysql.NewConfig()
	cfg.Net = "unix"
	cfg.Addr = strings.TrimPrefix(sock, "unix://")
	cfg.User = "qcloud"
	cfg.DBName = name
	return cfg
}

func (d *qDriver) mysqlClient() (pb.MySqlClient, error) {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	if d.client == nil {
		cc, err := qconn.Dial(*addr)
		if err != nil {
			return nil, fmt.Errorf("cannot connect to discovery: %v", err)
		}
		d.client = pb.NewMySqlClient(cc)
	}
	return d.client, nil
}

type dsnConnector struct {
	d    *qDriver
	name string
}

func (c *dsnConnector) Connect(ctx context.Context) (driver.Conn, error) {
	client, err := c.d.mysqlClient()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := client.GetDsn(ctx, &pb.GetDsnReq{Name: c.name})
	if err != nil {
		return nil, fmt.Errorf("cannot get DSN: %v", err)
	}

	return mysql.MySQLDriver{}.Open(resp.Dsn)
}

func (c *dsnConnector) Driver() driver.Driver { return c.d }
