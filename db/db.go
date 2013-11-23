/*
 This is the latest version of db

 Time-stamp: <[db.go] Elivoa @ Saturday, 2013-11-23 23:12:56>
*/
package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

// Connect create a connection to database
func Connect() (*sql.DB, error) {
	var err error
	conn, err := sql.Open("mysql", "root:eserver409$)(@/syd?charset=utf8&parseTime=true&loc=Local")
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// ConnectP create a connection to databse, it panics when any error occurs.
func Connectp() *sql.DB {
	conn, err := Connect()
	if err != nil {
		panic("Error when Connect to database: " + err.Error())
	}
	return conn
}

func CloseConn(conn *sql.DB) {
	conn.Close()
	DB.Close()
}

func Close() {
	DB.Close()
}

func CloseStmt(stmt *sql.Stmt) {
	if stmt != nil {
		stmt.Close()
	}
}

func CloseRows(rows *sql.Rows) {
	if rows != nil {
		rows.Close()
	}
}

/*
   params
*/

type Filter struct {
	// TODO design a filter/parameter
}

/*
  Error handling
*/
func Err(err error) bool {
	if err != nil {
		fmt.Println("xxxxxxxx  DB ERROR  xxxxxxxxxxxxxxxxxxxxxxxx")
		panic(err.Error())
		fmt.Println("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
		return true
	}
	return false
}
