package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/denisenkom/go-mssqldb"
)

var DB *sql.DB

func Connect() {

	connString := os.Getenv("DB_CONN")

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
