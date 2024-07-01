package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	_, err := sql.Open("sqlite3", "test.db")

	if err != nil {
		fmt.Println(err)
	}

	// if err := db.Ping(); err != nil {
	// fmt.Println(err)
	// }
}
