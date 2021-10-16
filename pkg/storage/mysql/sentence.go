package mysql

import (
	"database/sql"
	"errors"
	"strings"

	sq "github.com/Masterminds/squirrel"
	. "github.com/shubhamdwivedii/collab-story-assignment/pkg/sentence"
)

/************************ Sentence *****************************/

func (s *MySQLStorage) AddSentence(paragraphId int32, word string) (int32, error) {
	// check if paragraph exists
	paragraph, err := s.GetParagraph(paragraphId)
	if err != nil {
		return 0, err
	}

	tx, err := s.NewTransaction()
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
	if err := UpdateStoryUpdateTimeTx(tx, paragraph.Story); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, errors.New("Error Executing Transaction...")
	}

	return sentenceId, nil
}

func (s *MySQLStorage) GetSentence(sentenceId int32) (*Sentence, error) {
	var sentence Sentence

	query, args, err := sq.Select("*").From("sentences").Where(sq.Eq{"id": sentenceId}).ToSql()

	if err != nil {
		return nil, errors.New("Unexpected Error in Creating Query:" + err.Error())
	}

	var isFinished int

	if err := s.db.QueryRow(query, args...).Scan(
		&sentence.ID,
		&sentence.Paragraph,
		&isFinished,
		&sentence.Content,
	); err != nil {
		return nil, errors.New("Cannot Find Sentence IN DB:" + err.Error())
	}

	if isFinished == 1 {
		sentence.IsFinished = true
	}

	return &sentence, nil
}

func GetParagraphSentencesTx(tx *sql.Tx, paragraphId int32) ([]string, error) {
	var sentences []string

	query := sq.Select("content").From("sentences").Where(sq.Eq{"paragraph": paragraphId})

	rows, err := query.RunWith(tx).Query()

	if err != nil {
		// tx.Rollback()
		return nil, errors.New("Unexpected Error Creting query")
	}

	for rows.Next() {
		var sentence string
		err = rows.Scan(
			&sentence,
		)

		if err != nil {
			tx.Rollback()
			return nil, err
		}

		sentences = append(sentences, sentence)
	}

	return sentences, nil
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
	tx, err := s.NewTransaction()

	if err != nil {
		return err
	}

	sentence, err := GetSentenceTx(tx, sentenceId)

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

	if err := UpdateSentenceTx(tx, *sentence); err != nil {
		return err
	}

	if sentence.IsFinished {
		// Check if Paragraph is Finished ? 10 sentences
		// count = GetFinishedSentenceCount(paragraphId)
		count, err := CountFinishedSentencesTx(tx, sentence.Paragraph)
		if err != nil {
			// Tx already rolledback in countFinishedSentencesTx
			return err
		}

		if count >= 10 {
			// Check if paragraph is finished already or not, update isFinished
			paragraph, err := GetParagraphTx(tx, sentence.Paragraph)
			if err != nil {
				// tx already rolled back in getParagraphTx
				return err
			}

			if !paragraph.IsFinished {
				// Mark Paragraph as finished.
				paragraph.IsFinished = true
				if err := UpdateParagraphTx(tx, *paragraph); err != nil {
					// tx rolledback
					return err
				}
				// Now Check if Story is finished now or not ?
				count, err := CountFinishedParagraphsTx(tx, paragraph.Story)
				if err != nil {
					// tx already rolledback in countFinishedParagraphsTx
					return err
				}

				if count >= 7 {
					// Check if story is finished already or not, update isFinished.
					story, err := GetStoryTx(tx, paragraph.Story)
					if err != nil {
						return err
					}

					if !story.IsFinished {
						// Mark Story as Finished.
						story.IsFinished = true
						if err := UpdateStoryTx(tx, *story); err != nil {
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

func GetSentenceTx(tx *sql.Tx, sentenceId int32) (*Sentence, error) {
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

func UpdateSentenceTx(tx *sql.Tx, sentence Sentence) error {
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
