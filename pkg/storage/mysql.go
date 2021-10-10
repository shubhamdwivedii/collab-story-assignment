package storage

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/go-sql-driver/mysql"

	. "github.com/shubhamdwivedii/collab-story-assignment/pkg/paragraph"
	. "github.com/shubhamdwivedii/collab-story-assignment/pkg/sentence"
	. "github.com/shubhamdwivedii/collab-story-assignment/pkg/story"
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

func (s *MySQLStorage) newTransaction() (*sql.Tx, error) {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.New("Unexpected Error When Accessing DB..")
	}
	return tx, nil
}

/************************ Sentence *****************************/

// CREATE TABLE sentences (
//     id int NOT NULL AUTO_INCREMENT,
//     paragraph int NOT NULL,
//     isFinished tinyint(1) DEFAULT 0,
//     content varchar(240) DEFAULT '', -- 15 chars in 15 words + 14 spaces in between ~ 240
//     PRIMARY KEY (id)
// )

func (s *MySQLStorage) AddSentence(paragraphId int32) error {
	// check if paragraph exists
	paragraph, err := s.GetParagraph(paragraphId)
	if err != nil {
		return err
	}

	tx, err := s.newTransaction()
	if err != nil {
		return err
	}

	query := sq.Insert("sentences").Columns("paragraph").Values(paragraphId)
	// other values have default

	if _, err := query.RunWith(tx).Exec(); err != nil {
		return errors.New("Error Inserting Sentence To DB:" + err.Error())
	}

	// Update Story's UpdatedAt
	if err := updateStoryUpdateTimeTx(tx, paragraph.Story); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.New("Error Executing Transaction...")
	}

	return nil
}

/************************ Paragraph ******************************/

// CREATE TABLE paragraphs (
//     id int NOT NULL AUTO_INCREMENT,
//     story int NOT NULL,
//     isFinished tinyint(1) DEFAULT 0,
//     PRIMARY KEY (id)
// )

func (s *MySQLStorage) AddParagraph(storyId int32) error {
	// check if story exists
	story, err := s.GetStory(storyId)
	if err != nil {
		return err
	}

	tx, err := s.newTransaction()
	if err != nil {
		return err // error already formatted in newTransaction
	}

	query := sq.Insert("paragraphs").Columns("story").Values(storyId)
	// other values have default

	if _, err := query.RunWith(tx).Exec(); err != nil {
		return errors.New("Error Inserting Paragraph To DB:" + err.Error())
	}

	// Update Story's UpdatedAt
	if err := updateStoryUpdateTimeTx(tx, story.ID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.New("Error Executing Transaction...")
	}

	return nil
}

func (s *MySQLStorage) GetParagraph(paragraphId int32) (*Paragraph, error) {
	var paragraph Paragraph

	query, args, err := sq.Select("*").From("paragraphs").Where(sq.Eq{"id": paragraphId}).ToSql()

	if err != nil {
		return nil, errors.New("Unexpected Error In Creating Query:" + err.Error())
	}

	var isFinished int

	if err := s.db.QueryRow(query, args...).Scan(
		&paragraph.ID,
		&paragraph.Story,
		&isFinished,
	); err != nil {
		return nil, errors.New("Cannot Find Paragraph In DB:" + err.Error())
	}

	if isFinished == 1 {
		paragraph.IsFinished = true
	}
	return &paragraph, nil
}

/************************** Story ********************************/

func (s *MySQLStorage) AddStory() error {
	log.Println("Adding Story")

	query := sq.Insert("stories").Columns().Values()
	// All values have defaults.

	if _, err := query.RunWith(s.db).Exec(); err != nil {
		return errors.New("Error Adding Story Into DB.." + err.Error())
	}
	return nil
}

func (s *MySQLStorage) GetStory(storyId int32) (*Story, error) {
	var story Story

	query, args, err := sq.Select("*").From("stories").Where(sq.Eq{"id": storyId}).ToSql()

	if err != nil {
		return nil, errors.New("Unexpected Error In Creating Query:" + err.Error())
	}

	var titleAdded, isFinished int

	if err := s.db.QueryRow(query, args...).Scan(
		&story.ID,
		&story.Title,
		&titleAdded,
		&isFinished,
		&story.CreatedAt,
		&story.UpdatedAt,
	); err != nil {
		return nil, errors.New("Cannot Find Story IN DB:" + err.Error())
	}

	if titleAdded == 1 {
		story.TitleAdded = true
	}
	if isFinished == 1 {
		story.IsFinished = true
	}

	return &story, nil
}

func updateStoryTx(tx *sql.Tx, story Story) error {
	query, args, err := sq.Update("stories").
		Set("title", story.Title).
		Set("titleAdded", story.TitleAdded).
		Set("isFinished", story.IsFinished).
		Set("updatedAt", story.UpdatedAt.Format(MySQLTimeFormat)).
		Where(sq.Eq{"id": story.ID}).ToSql()

	if err != nil {
		return errors.New("Error Generating Story Update Query:" + err.Error())
	}

	if _, err = tx.Exec(query, args...); err != nil {
		tx.Rollback()
		return errors.New("Error Updating Story In DB:" + err.Error())
	}
	return nil
}

func updateStoryUpdateTimeTx(tx *sql.Tx, storyId int32) error {
	query, args, err := sq.Update("stories").
		Set("updatedAt", time.Now().Format(MySQLTimeFormat)).
		Where(sq.Eq{"id": storyId}).ToSql()

	if err != nil {
		return errors.New("Error Generating Story Update Query:" + err.Error())
	}

	if _, err = tx.Exec(query, args...); err != nil {
		tx.Rollback()
		return errors.New("Error Updating Story In DB:" + err.Error())
	}

	return nil
}
