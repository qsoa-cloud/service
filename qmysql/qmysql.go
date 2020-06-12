//go:generate protoc -I internal/pb --go_out=plugins=grpc:internal/pb internal/pb/qmysql.proto
package qmysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"flag"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"

	"gopkg.qsoa.cloud/service/qmysql/internal/pb"
)

type qDriver struct {
	client pb.MySqlClient
}

var (
	addr = flag.String("q_mysql_addr", "", "mysql discovery address")
)

func init() {
	sql.Register("qmysql", &qDriver{})
}

func (d *qDriver) Open(name string) (driver.Conn, error) {
	if d.client == nil {
		flag.Parse()

		cc, err := grpc.Dial(*addr, grpc.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("cannot connect to discovery: %v", err)
		}

		d.client = pb.NewMySqlClient(cc)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := d.client.GetDsn(ctx, &pb.GetDsnReq{Name: name})
	if err != nil {
		return nil, fmt.Errorf("cannot get DSN: %v", err)
	}

	return mysql.MySQLDriver{}.Open(resp.Dsn)
}
