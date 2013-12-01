/*
 This is the latest version of db

 Time-stamp: <[db.go] Elivoa @ Sunday, 2013-12-01 13:11:00>
*/
package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

var logdebug bool = false
var connections int = 0

// Connect create a connection to database
func Connect() (*sql.DB, error) {
	if logdebug {
		connections += 1
		fmt.Printf("^^^^^^^^^  db.Connect(), [%d] connections in total.\n", connections)
	}
	conn, err := sql.Open("mysql", "root:eserver409$)(@/syd?charset=utf8&parseTime=true&loc=Local&timeout=30s")
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
	fmt.Println(connections)
	if logdebug {
		connections -= 1
		fmt.Printf("vvvvvvvv  db.CloseConn(), [%d] connections left.\n", connections)
	}
	err := conn.Close()
	if err != nil {
		panic("Error when closing Connection to db. " + err.Error())
	}
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
