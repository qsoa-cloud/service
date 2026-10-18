package qmysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"

	"gopkg.qsoa.cloud/service/qmysql/internal/pb"
)

func TestProxyConfig(t *testing.T) {
	c := proxyConfig("unix:///qcloud/mysql_proxy.sock", "orders")
	if c.Net != "unix" || c.Addr != "/qcloud/mysql_proxy.sock" || c.DBName != "orders" || c.Passwd != "" || c.User != "qcloud" {
		t.Fatalf("net %q addr %q db %q user %q, password set %v", c.Net, c.Addr, c.DBName, c.User, c.Passwd != "")
	}
}

func TestOpenDialsProxy(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "proxy.sock")
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	dialed := make(chan struct{}, 1)
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			select {
			case dialed <- struct{}{}:
			default:
			}
			c.Close()
		}
	}()
	*sockAddr = "unix://" + sock
	t.Cleanup(func() { *sockAddr = "" })

	db, err := sql.Open("qmysql", "orders")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err == nil {
		t.Fatal("ping succeeded")
	}
	select {
	case <-dialed:
	case <-time.After(5 * time.Second):
		t.Fatal("proxy socket was not dialed")
	}
}

type fakeDsn struct {
	pb.UnimplementedMySqlServer
	dsn   string
	names chan string
}

func (f *fakeDsn) GetDsn(_ context.Context, r *pb.GetDsnReq) (*pb.GetDsnResp, error) {
	f.names <- r.Name
	return &pb.GetDsnResp{Dsn: f.dsn}, nil
}

func TestConnectorUsesDsn(t *testing.T) {
	dir := t.TempDir()
	sock := filepath.Join(dir, "mysql.sock")
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	f := &fakeDsn{dsn: "u:hunter2@unix(" + filepath.Join(dir, "none.sock") + ")/orders", names: make(chan string, 1)}
	s := grpc.NewServer()
	pb.RegisterMySqlServer(s, f)
	go s.Serve(l)
	t.Cleanup(s.Stop)
	*addr = "unix://" + sock
	t.Cleanup(func() { *addr = "" })

	d := &qDriver{}
	c, err := driver.DriverContext(d).OpenConnector("orders")
	if err != nil {
		t.Fatal(err)
	}
	if c.Driver() != d {
		t.Fatal("connector driver is not qmysql")
	}
	if _, err := c.Connect(context.Background()); err == nil || strings.Contains(err.Error(), "hunter2") {
		t.Fatalf("Connect: %v", err)
	}
	if n := <-f.names; n != "orders" {
		t.Fatalf("GetDsn name %q", n)
	}
}
