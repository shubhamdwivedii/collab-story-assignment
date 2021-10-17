package mysql

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

const MySQLTimeFormat = "2006-01-02 15:04:05"

type MySQLStorage struct {
	db     *sql.DB
	logger *log.Logger
}

func NewMySQLStorage(connection string, logger *log.Logger) (*MySQLStorage, error) {
	var err error
	s := new(MySQLStorage)
	if s.db, err = initDb(connection); err != nil {
		logger.Error("Error Connection to DB...", err)
		return nil, err
	}
	logger.Info("Connected to DB successfully...")
	s.logger = logger
	return s, nil
}

func (s *MySQLStorage) NewTransaction() (*sql.Tx, error) {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.New("Unexpected Error When Accessing DB..")
	}
	return tx, nil
}

func initDb(connection string) (*sql.DB, error) {
	db, err := sql.Open("mysql", connection+"?parseTime=true")
	// adding ?parseTime=true will parse sql's DATETIME to time.Time when scanning.

	if err != nil {
		return nil, err
	}
	return db, db.Ping()
}
