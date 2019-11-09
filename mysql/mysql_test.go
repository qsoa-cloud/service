package mysql_test

import (
	"database/sql"
	"flag"
	"testing"

	_ "gopkg.qsoa.cloud/service/mysql"
)

var dbName = flag.String("dbname", "", "")

func TestQDriver_Open(t *testing.T) {
	if *dbName == "" {
		return
	}

	c, err := sql.Open("qmysql", *dbName)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	if err := c.Ping(); err != nil {
		t.Fatal(err)
	}
}
