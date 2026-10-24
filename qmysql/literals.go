package qmysql

import "gopkg.qsoa.cloud/service/internal/sqltext"

func stripLiterals(sql string) string { return sqltext.Strip(sql, sqltext.MySQL) }
