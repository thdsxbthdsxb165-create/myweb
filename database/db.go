package database

import (
	"database/sql"
	"log"

	_ "github.com/denisenkom/go-mssqldb"
)

var DB *sql.DB

func Connect() {
	connString := "sqlserver://sa:123456@127.0.0.1:1433?database=myweb"

	var err error
	DB, err = sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal(err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MSSQL")
}
