// Copyright (c) , zhoucb, Strong Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/zhouchangbo/ss-go

package db2mysql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"runtime"
	"sslog"
)

func GetConnection(sqlUrl string) (*sql.DB, error) {
	_, file, line, _ := runtime.Caller(0)

	db, err := sql.Open("mysql", sqlUrl)
	if err != nil {
		sslog.LoggerErr("[%s:%d]sql.Open: err=%s", file, line, err.Error())
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		sslog.LoggerErr("[%s:%d]db.Ping: err=%s", file, line, err.Error())
		return nil, err
	}

	return db, err
}

func SelectSql(db *sql.DB, sql string) (*sql.Rows, error) {
	rows, err := db.Query(sql)
	return rows, err
}

func DbClose(db *sql.DB) {
	db.Close()
}

func RowsClose(rows *sql.Rows) {
	rows.Close()
}
