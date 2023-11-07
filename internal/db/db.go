package db

import (
	"database/sql"
	"os"
)

var DB *sql.DB

func InitDB() {
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))

	if err != nil {
		panic(err)
	}

	DB = db
}
