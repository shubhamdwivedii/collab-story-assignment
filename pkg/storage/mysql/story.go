package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	sq "github.com/Masterminds/squirrel"
	. "github.com/shubhamdwivedii/collab-story-assignment/pkg/story"
)

/************************** Story ********************************/

func (s *MySQLStorage) AddStory() (int32, error) {
	log.Println("Adding Story")

	query := sq.Insert("stories").Columns().Values()
	// All values have defaults.

	if res, err := query.RunWith(s.db).Exec(); err != nil {
		return 0, errors.New("Error Adding Story Into DB.." + err.Error())
	} else {
		id, _ := res.LastInsertId()
		return int32(id), nil
	}
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

func (s *MySQLStorage) GetStoryDetail(storyId int32) (*StoryResponse, error) {
	tx, err := s.NewTransaction()

	if err != nil {
		return nil, err
	}

	story, err := GetStoryTx(tx, storyId)

	if err != nil {
		return nil, err
	}

	paragraphs, err := GetStoryParagraphsTx(tx, storyId)

	if err != nil {
		return nil, err
	}

	// iterate through paragraph ids and get all sentences content.
	var paraBriefs []ParagraphBrief

	for _, para := range paragraphs {
		sentences, err := GetParagraphSentencesTx(tx, para.ID)
		if err != nil {
			return nil, err
		}
		paragraph := ParagraphBrief{
			ID:        para.ID,
			Sentences: sentences,
		}

		paraBriefs = append(paraBriefs, paragraph)
	}

	storyRes := StoryResponse{
		ID:         story.ID,
		Title:      story.Title,
		CreatedAt:  story.CreatedAt,
		UpdatedAt:  story.UpdatedAt,
		Paragraphs: paraBriefs,
	}

	return &storyRes, nil
}

func (s *MySQLStorage) UpdateStoryTitle(storyid int32, word string) error {
	tx, err := s.NewTransaction()

	if err != nil {
		return err // error already formatted in newTransaction
	}

	story, err := GetStoryTx(tx, storyid)

	if err != nil {
		return err // tx already rolledback in getStoryTx
	}

	// check if title already contains two words
	re, _ := regexp.Compile("^\\S+(?: \\S+)$")
	if re.MatchString(story.Title) || story.TitleAdded {
		return errors.New("Error: Story Title Is Finished Already")
	} else {
		if len(story.Title) == 0 {
			story.Title = word
		} else {
			story.Title = story.Title + " " + word
			story.TitleAdded = true
		}
	}

	if err := UpdateStoryTx(tx, *story); err != nil {
		return err
	}

	// Commit the Tx ???

	if err := tx.Commit(); err != nil {
		return errors.New("Errors Executing Transaction...")
	}

	return nil
}

func (s *MySQLStorage) GetAllStories(limit int32, offset int32) (*StoriesResponse, error) {
	tx, err := s.NewTransaction()

	if err != nil {
		return nil, err
	}

	var stories []StoryBrief

	query := sq.Select("*").From("stories").Limit(uint64(limit)).Offset(uint64(offset))

	rows, err := query.RunWith(tx).Query()

	if err != nil {
		// tx.Rollback()
		return nil, err
	}

	for rows.Next() {
		var story StoryBrief
		var isFinished, titleAdded int32
		err = rows.Scan(
			&story.ID,
			&story.Title,
			&titleAdded,
			&isFinished,
			&story.CreatedAt,
			&story.UpdatedAt,
		)

		if err != nil {
			tx.Rollback()
			return nil, err
		}

		fmt.Println("appending story", story)
		stories = append(stories, story)
	}

	qry, args, err := sq.Select("count(*) as count").From("stories").ToSql()

	if err != nil {
		tx.Rollback()
		return nil, errors.New("Cannof Build Query for Count")
	}

	var count int32
	if err := tx.QueryRow(qry, args...).Scan(
		&count,
	); err != nil {
		tx.Rollback()
		return nil, errors.New("Error getting count")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.New("Error Executing Transaction...")
	} else {
		storyRes := StoriesResponse{
			Limit:   limit,
			Offset:  offset,
			Count:   count,
			Results: stories,
		}
		fmt.Println("StoryRes", storyRes)
		return &storyRes, nil
	}
}

func (s *MySQLStorage) GetUnfinishedStory() (*Story, error) {
	var story Story

	query, args, err := sq.Select("*").From("stories").
		Where(sq.Eq{"isFinished": 0}).ToSql()

	if err != nil {
		return nil, errors.New("Unexpected Error In Creating Query:" + err.Error())
	}

	var titleAdded, isFinished int32
	if err := s.db.QueryRow(query, args...).Scan(
		&story.ID,
		&story.Title,
		&titleAdded,
		&isFinished,
		&story.CreatedAt,
		&story.UpdatedAt,
	); err != nil {
		return nil, errors.New("Cannot Find An Empty Sentence in DB:" + err.Error())
	}

	if titleAdded == 1 {
		story.TitleAdded = true
	}

	if isFinished == 1 {
		story.IsFinished = true
	}

	return &story, nil
}

func UpdateStoryTx(tx *sql.Tx, story Story) error {
	var titleAdded, isFinished int32
	if story.TitleAdded {
		titleAdded = 1
	}
	if story.IsFinished {
		isFinished = 1
	}
	query, args, err := sq.Update("stories").
		Set("title", story.Title).
		Set("titleAdded", titleAdded).
		Set("isFinished", isFinished).
		Set("updatedAt", time.Now().Format(MySQLTimeFormat)).
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

func UpdateStoryUpdateTimeTx(tx *sql.Tx, storyId int32) error {
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

func GetStoryTx(tx *sql.Tx, storyId int32) (*Story, error) {
	var story Story

	query, args, err := sq.Select("*").From("stories").Where(sq.Eq{"id": storyId}).ToSql()

	if err != nil {
		tx.Rollback()
		return nil, errors.New("Unexpected Error In Creating Query:" + err.Error())
	}

	var titleAdded, isFinished int

	if err := tx.QueryRow(query, args...).Scan(
		&story.ID,
		&story.Title,
		&titleAdded,
		&isFinished,
		&story.CreatedAt,
		&story.UpdatedAt,
	); err != nil {
		tx.Rollback()
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

func CountFinishedParagraphsTx(tx *sql.Tx, storyId int32) (int32, error) {
	query, args, err := sq.Select("count(*) as count").From("paragraphs").
		Where(sq.Eq{"story": storyId}, sq.Eq{"isFinished": 1}).ToSql()

	if err != nil {
		tx.Rollback()
		return 0, errors.New("Cannot Find Count of Finished Paragraphs" + err.Error())
	}

	var count int32
	if err := tx.QueryRow(query, args...).Scan(
		&count,
	); err != nil {
		tx.Rollback()
		return 0, errors.New("Cannot Find Count Of Paragraphs" + err.Error())
	}

	return count, nil
}