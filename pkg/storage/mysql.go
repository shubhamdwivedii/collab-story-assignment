package storage

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"regexp"
	"strings"
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

func (s *MySQLStorage) AddSentence(paragraphId int32, word string) (int32, error) {
	// check if paragraph exists
	paragraph, err := s.GetParagraph(paragraphId)
	if err != nil {
		return 0, err
	}

	tx, err := s.newTransaction()
	if err != nil {
		return 0, err
	}

	query := sq.Insert("sentences").Columns("paragraph", "content").Values(paragraphId, word)
	// other values have default

	var sentenceId int32
	if res, err := query.RunWith(tx).Exec(); err != nil {
		return 0, errors.New("Error Inserting Sentence To DB:" + err.Error())
	} else {
		id, _ := res.LastInsertId()
		sentenceId = int32(id)
	}

	// Update Story's UpdatedAt
	if err := updateStoryUpdateTimeTx(tx, paragraph.Story); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, errors.New("Error Executing Transaction...")
	}

	return sentenceId, nil
}

func (s *MySQLStorage) GetUnfinishedSentence(paragraphId int32) (*Sentence, error) {
	var sentence Sentence

	query, args, err := sq.Select("*").From("sentences").
		Where(sq.Eq{"isFinished": 0}, sq.Eq{"paragraph": paragraphId}).ToSql()

	if err != nil {
		return nil, errors.New("Unexpected Error In Creating Query:" + err.Error())
	}

	var isFinished int32
	if err := s.db.QueryRow(query, args...).Scan(
		&sentence.ID,
		&sentence.Paragraph,
		&isFinished,
		&sentence.Content,
	); err != nil {
		return nil, errors.New("Cannot Find An Empty Sentence in DB:" + err.Error())
	}

	if isFinished == 1 {
		sentence.IsFinished = true
	}

	return &sentence, nil
}

func (s *MySQLStorage) UpdateSentence(sentenceId int32, word string) error {
	tx, err := s.newTransaction()

	if err != nil {
		return err
	}

	sentence, err := getSentenceTx(tx, sentenceId)

	if err != nil {
		return err // rollback already done in getSentenceTx
	}

	// check if sentence is already finished
	words := strings.Split(sentence.Content, " ")
	if len(words) >= 15 || sentence.IsFinished {
		return errors.New("Error: Sentence Is Already Finished")
	} else {
		// Not Finished, Add one more Word.
		sentence.Content = sentence.Content + " " + word
		if len(words)+1 == 15 { // Is it finished now ?
			sentence.IsFinished = true
		}
	}

	if err := updateSentenceTx(tx, *sentence); err != nil {
		return err
	}

	if sentence.IsFinished {
		// Check if Paragraph is Finished ? 10 sentences
		// count = GetFinishedSentenceCount(paragraphId)
		count, err := countFinishedSentencesTx(tx, sentence.Paragraph)
		if err != nil {
			// Tx already rolledback in countFinishedSentencesTx
			return err
		}

		if count >= 10 {
			// Check if paragraph is finished already or not, update isFinished
			paragraph, err := getParagraphTx(tx, sentence.Paragraph)
			if err != nil {
				// tx already rolled back in getParagraphTx
				return err
			}

			if !paragraph.IsFinished {
				// Mark Paragraph as finished.
				paragraph.IsFinished = true
				if err := updateParagraphTx(tx, *paragraph); err != nil {
					// tx rolledback
					return err
				}
				// Now Check if Story is finished now or not ?
				count, err := countFinishedParagraphsTx(tx, paragraph.Story)
				if err != nil {
					// tx already rolledback in countFinishedParagraphsTx
					return err
				}

				if count >= 7 {
					// Check if story is finished already or not, update isFinished.
					story, err := getStoryTx(tx, paragraph.Story)
					if err != nil {
						return err
					}

					if !story.IsFinished {
						// Mark Story as Finished.
						story.IsFinished = true
						if err := updateStoryTx(tx, *story); err != nil {
							// tx rolledback
							return err
						}
					}
				} else {
					// Nothing Finished Paragraphs is still less than 10.
				}
			}
		} else {
			// Nothing Finished Sentences is still less than 10.
		}
	} else {
		// Nothing Sentence can have more words
	}

	if err := tx.Commit(); err != nil {
		return errors.New("Errors Executing Transaction...")
	}

	return nil
}

func getSentenceTx(tx *sql.Tx, sentenceId int32) (*Sentence, error) {
	var sentence Sentence

	query, args, err := sq.Select("*").From("sentences").Where(sq.Eq{"id": sentenceId}).ToSql()

	if err != nil {
		tx.Rollback()
		return nil, errors.New("Unexpected Error in Creating Query:" + err.Error())
	}

	var isFinished int

	if err := tx.QueryRow(query, args...).Scan(
		&sentence.ID,
		&sentence.Paragraph,
		&isFinished,
		&sentence.Content,
	); err != nil {
		tx.Rollback()
		return nil, errors.New("Cannot Find Sentence IN DB:" + err.Error())
	}

	if isFinished == 1 {
		sentence.IsFinished = true
	}

	return &sentence, nil
}

func updateSentenceTx(tx *sql.Tx, sentence Sentence) error {
	var isFinished int32
	if sentence.IsFinished {
		isFinished = 1
	}
	query, args, err := sq.Update("sentences").
		Set("content", sentence.Content).
		Set("isFinished", isFinished).
		Where(sq.Eq{"id": sentence.ID}).ToSql()

	if err != nil {
		return errors.New("Error Generating Sentence Update Query:" + err.Error())
	}

	if _, err = tx.Exec(query, args...); err != nil {
		tx.Rollback()
		return errors.New("Error Updating Sentence In DB:" + err.Error())
	}
	return nil
}

/************************ Paragraph ******************************/

func (s *MySQLStorage) AddParagraph(storyId int32) (int32, error) {
	// check if story exists
	story, err := s.GetStory(storyId)
	if err != nil {
		return 0, err
	}

	tx, err := s.newTransaction()
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
	if err := updateStoryUpdateTimeTx(tx, story.ID); err != nil {
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

func countFinishedSentencesTx(tx *sql.Tx, paragraphId int32) (int32, error) {
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

func getParagraphTx(tx *sql.Tx, paragraphId int32) (*Paragraph, error) {
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

func updateParagraphTx(tx *sql.Tx, paragraph Paragraph) error {
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

func (s *MySQLStorage) UpdateStoryTitle(storyid int32, word string) error {
	tx, err := s.newTransaction()

	if err != nil {
		return err // error already formatted in newTransaction
	}

	story, err := getStoryTx(tx, storyid)

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

	if err := updateStoryTx(tx, *story); err != nil {
		return err
	}

	// Commit the Tx ???

	if err := tx.Commit(); err != nil {
		return errors.New("Errors Executing Transaction...")
	}

	return nil
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

func updateStoryTx(tx *sql.Tx, story Story) error {
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

func getStoryTx(tx *sql.Tx, storyId int32) (*Story, error) {
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

func countFinishedParagraphsTx(tx *sql.Tx, storyId int32) (int32, error) {
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
