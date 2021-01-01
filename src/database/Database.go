package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"
	"visual-feed-aggregator/src/util/logging"

	"github.com/jmoiron/sqlx"
)

// OpenDbWithSchema opens up a new connection and makes sure to create the database & schema, in case they dont exist
func OpenDbWithSchema(user, pass, address string) (*sqlx.DB, error) {
	var err error
	for retries := 3; retries > 0; retries-- {
		_, err = tryCreateDatabase(user, pass, address)
		if err == nil {
			break
		} else {
			logging.Println(logging.Warn, "Error trying to create DB, will retry shortly...")
			time.Sleep(30 * time.Second)
		}
	}
	if err != nil {
		return nil, err
	}
	err = tryCreateShema(user, pass, address)
	if err != nil {
		return nil, err
	}

	return sqlx.Connect("mysql", fmt.Sprintf("%s:%s@tcp(%s)/vifa?tls=preferred&parseTime=true", user, pass, address))
}

// tryCreateDatabase creates a database if it does not exist and returns true if it had to create it
func tryCreateDatabase(user, pass, address string) (bool, error) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := sqlx.ConnectContext(timeoutCtx, "mysql", fmt.Sprintf("%s:%s@tcp(%s)/", user, pass, address))
	if err != nil {
		return false, err
	}
	defer db.Close()

	rows, err := db.Query("SHOW DATABASES LIKE 'vifa';")
	if err != nil {
		return false, err
	}
	createdDB := !rows.Next()
	rows.Close()

	if createdDB {
		_, err = db.Exec("CREATE DATABASE IF NOT EXISTS vifa;")
		if err != nil {
			return true, err
		}
	}

	return createdDB, nil
}

func tryCreateShema(user, pass, address string) error {
	schemaFn := "." + string(os.PathSeparator) + path.Join("res", "database", "schema.sql")
	schemaBytes, err := ioutil.ReadFile(schemaFn)
	schema := string(schemaBytes)

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/vifa?multiStatements=true", user, pass, address))
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(schema)
	if err != nil {
		return err
	}

	return nil
}
