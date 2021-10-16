package mysql

import (
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	. "github.com/shubhamdwivedii/collab-story-assignment/pkg/paragraph"
)

/************************ Paragraph ******************************/

func (s *MySQLStorage) AddParagraph(storyId int32) (int32, error) {
	// check if story exists
	story, err := s.GetStory(storyId)
	if err != nil {
		return 0, err
	}

	tx, err := s.NewTransaction()
	if err != nil {
		return 0, err // error already formatted in newTransaction
	}

	query := sq.Insert("paragraphs").Columns("story").Values(storyId)
	// other values have default

	var paragraphId int32
	if res, err := query.RunWith(tx).Exec(); err != nil {
		return 0, errors.New("Error Inserting Paragraph To DB:" + err.Error())
	} else {
		id, _ := res.LastInsertId()
		paragraphId = int32(id)
	}

	// Update Story's UpdatedAt
	if err := UpdateStoryUpdateTimeTx(tx, story.ID); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, errors.New("Error Executing Transaction...")
	}

	return paragraphId, nil
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

func GetStoryParagraphsTx(tx *sql.Tx, storyId int32) ([]Paragraph, error) {
	var paragraphs []Paragraph

	query := sq.Select("*").From("paragraphs").Where(sq.Eq{"story": storyId})

	rows, err := query.RunWith(tx).Query()

	if err != nil {
		// tx.Rollback()
		return nil, errors.New("Unexpected Error in Creating Query:" + err.Error())
	}

	for rows.Next() {
		var paragraph Paragraph
		var isFinished int32
		err = rows.Scan(
			&paragraph.ID,
			&paragraph.Story,
			&isFinished,
		)

		if err != nil {
			tx.Rollback()
			return nil, err
		}

		paragraphs = append(paragraphs, paragraph)
	}

	return paragraphs, nil
}

func (s *MySQLStorage) GetUnfinishedParagraph(storyId int32) (*Paragraph, error) {
	var paragraph Paragraph

	query, args, err := sq.Select("*").From("paragraphs").Where(sq.Eq{"isFinished": 0}, sq.Eq{"story": storyId}).ToSql()

	if err != nil {
		return nil, errors.New("Unexpected Error In Creating Query:" + err.Error())
	}

	var isFinished int32
	if err := s.db.QueryRow(query, args...).Scan(
		&paragraph.ID,
		&paragraph.Story,
		&isFinished,
	); err != nil {
		return nil, errors.New("Cannot Find An Empty Sentence in DB:" + err.Error())
	}

	if isFinished == 1 {
		paragraph.IsFinished = true
	}

	return &paragraph, nil
}

func CountFinishedSentencesTx(tx *sql.Tx, paragraphId int32) (int32, error) {
	query, args, err := sq.Select("count(*) as count").From("sentences").
		Where(sq.Eq{"paragraph": paragraphId}, sq.Eq{"isFinished": 1}).ToSql()

	if err != nil {
		tx.Rollback()
		return 0, errors.New("Cannot Find Count of Finished Sentences" + err.Error())
	}

	var count int32
	if err := tx.QueryRow(query, args...).Scan(
		&count,
	); err != nil {
		tx.Rollback()
		return 0, errors.New("Cannot Find Count Of Sentences" + err.Error())
	}

	return count, nil
}

func GetParagraphTx(tx *sql.Tx, paragraphId int32) (*Paragraph, error) {
	var paragraph Paragraph

	query, args, err := sq.Select("*").From("paragraphs").Where(sq.Eq{"id": paragraphId}).ToSql()

	if err != nil {
		tx.Rollback()
		return nil, errors.New("Unexpected Error in Creating Query:" + err.Error())
	}

	var isFinished int
	if err := tx.QueryRow(query, args...).Scan(
		&paragraph.ID,
		&paragraph.Story,
		&isFinished,
	); err != nil {
		tx.Rollback()
		return nil, errors.New("Cannot Find Paragraph In DB:" + err.Error())
	}

	if isFinished == 1 {
		paragraph.IsFinished = true
	}

	return &paragraph, nil
}

func UpdateParagraphTx(tx *sql.Tx, paragraph Paragraph) error {
	var isFinished int32
	if paragraph.IsFinished {
		isFinished = 1
	}

	query, args, err := sq.Update("paragraphs").
		Set("isFinished", isFinished).
		Where(sq.Eq{"id": paragraph.ID}).ToSql()

	if err != nil {
		return errors.New("Error Generating Paragraph Update Query:" + err.Error())
	}

	if _, err = tx.Exec(query, args...); err != nil {
		tx.Rollback()
		return errors.New("Error Updating Paragraph in DB:" + err.Error())
	}
	return nil
}
