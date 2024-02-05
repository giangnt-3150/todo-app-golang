package db

import (
	"database/sql"
	"fmt"
	"log"
)

func ConnectDB() *sql.DB {
	username := "todoapp"
	password := "root"
	host := "mysql"
	port := "3306"
	databaseName := "todoapp"

	// Create the data source name (DSN)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, databaseName)

	// Open a connection to the database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	// Ping the database to check the connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	return db
}
