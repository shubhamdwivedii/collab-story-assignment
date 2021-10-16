package mysql

import (
	"context"
	"database/sql"
	"errors"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

const MySQLTimeFormat = "2006-01-02 15:04:05"

type MySQLStorage struct {
	db *sql.DB
}

func NewMySQLStorage(connection string) (*MySQLStorage, error) {
	var err error
	s := new(MySQLStorage)
	if s.db, err = initDb(connection); err != nil {
		return nil, err
	}
	return s, nil
}

func initDb(connection string) (*sql.DB, error) {
	db, err := sql.Open("mysql", connection+"?parseTime=true")
	// adding ?parseTime=true will parse sql's DATETIME to time.Time when scanning.

	if err != nil {
		log.Println("Unable to open connection to DB...", err.Error())
	}
	log.Println("Connected to DB successfully...")
	return db, db.Ping()
}

func (s *MySQLStorage) NewTransaction() (*sql.Tx, error) {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.New("Unexpected Error When Accessing DB..")
	}
	return tx, nil
}
