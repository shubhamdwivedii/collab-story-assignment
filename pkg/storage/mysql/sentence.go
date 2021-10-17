package mysql

import (
	"database/sql"
	"errors"
	"strings"

	sq "github.com/Masterminds/squirrel"
	. "github.com/shubhamdwivedii/collab-story/pkg/sentence"
)

// Add a new Sentence to DB
func (s *MySQLStorage) AddSentence(paragraphId int32, word string) (int32, error) {
	// Check if paragraph exists
	tx, err := s.NewTransaction()

	if err != nil {
		s.logger.Error(err)
		return 0, err // err already formatted in NewTransaction
	}

	paragraph, err := GetParagraphTx(tx, paragraphId)
	if err != nil {
		s.logger.Error(err)
		return 0, err // err already formatted in GetParagraphTx
	}

	query := sq.Insert("sentences").Columns("paragraph", "content").Values(paragraphId, word)
	// other values have default

	var sentenceId int32
	if res, err := query.RunWith(tx).Exec(); err != nil {
		tx.Rollback()
		s.logger.Error("Error Adding Sentence To DB:" + err.Error())
		// Abstract DB error messages.
		return 0, errors.New("Error Inserting Sentence To DB...")
	} else {
		id, _ := res.LastInsertId()
		sentenceId = int32(id)
	}

	// Update Story's UpdatedAt
	if err := UpdateStoryUpdateTimeTx(tx, paragraph.Story); err != nil {
		tx.Rollback()
		s.logger.Error("Error Updating Story's UpdatedAt:" + err.Error())
		return 0, errors.New("Error Updating Story's UpdatedAt...")
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("Error Commiting Transaction:" + err.Error())
		return 0, errors.New("Error Commiting Transaction...")
	}

	return sentenceId, nil
}

func (s *MySQLStorage) GetSentence(sentenceId int32) (*Sentence, error) {
	tx, err := s.NewTransaction()

	if err != nil {
		s.logger.Error(err) // err already formatted in NewTransaction
		return nil, err
	}

	// Check if Sentence exists.
	sentence, err := GetSentenceTx(tx, sentenceId)
	if err != nil {
		s.logger.Error(err) // err already formatted in GetSentenceTx
		return nil, err     // rollback already done in getSentenceTx
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("Error executing transaction:" + err.Error())
		return nil, errors.New("Errors executing transaction...")
	}

	return sentence, nil
}

func (s *MySQLStorage) GetUnfinishedSentence(paragraphId int32) (*Sentence, error) {
	var sentence Sentence

	query, args, err := sq.Select("*").From("sentences").
		Where(sq.Eq{"isFinished": 0}, sq.Eq{"paragraph": paragraphId}).ToSql()

	if err != nil {
		s.logger.Error("Unexpected Error In Creating Query:" + err.Error())
		return nil, errors.New("Unexpected Error In Creating Query...")
	}

	var isFinished int32
	if err := s.db.QueryRow(query, args...).Scan(
		&sentence.ID,
		&sentence.Paragraph,
		&isFinished,
		&sentence.Content,
	); err != nil {
		s.logger.Info("Cannot Find An Unfinished Sentence in DB:" + err.Error())
		return nil, errors.New("Cannot Find An Unfinished Sentence in DB...")
	}

	if isFinished == 1 {
		sentence.IsFinished = true
	}

	return &sentence, nil
}

// Updates a Sentence's Content (Words)
func (s *MySQLStorage) UpdateSentence(sentenceId int32, word string) error {
	tx, err := s.NewTransaction()

	if err != nil {
		s.logger.Error(err) // err already formatted in NewTransaction
		return err
	}

	// Check if Sentence exists.
	sentence, err := GetSentenceTx(tx, sentenceId)
	if err != nil {
		s.logger.Error(err) // err already formatted in GetSentenceTx
		return err          // rollback already done in getSentenceTx
	}

	// Check if Sentence is already finished
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

	// Update Sentence's Content or IsFinished status.
	if err := UpdateSentenceTx(tx, *sentence); err != nil {
		s.logger.Error(err) // err already formatted in UpdateSentenceTx
		return err          // Already rolledback in UpdateSentenceTx
	}

	// If Sentence was JUST marked finished (with the last word added)
	if sentence.IsFinished {
		// Check if Paragraph is Finished (has 10 sentences now)
		count, err := CountFinishedSentencesTx(tx, sentence.Paragraph)
		if err != nil {
			s.logger.Error(err)
			// Tx already rolledback in countFinishedSentencesTx
			return err
		}

		if count >= 10 {
			// Check if Paragraph is finished already or not then update isFinished status.
			paragraph, err := GetParagraphTx(tx, sentence.Paragraph)
			if err != nil {
				s.logger.Error(err)
				// tx already rolled back in getParagraphTx
				return err
			}

			if !paragraph.IsFinished {
				// Mark Paragraph as finished.
				paragraph.IsFinished = true
				if err := UpdateParagraphTx(tx, *paragraph); err != nil {
					s.logger.Error(err) // err already formatted
					// tx rolledback already.
					return err
				}

				// Now Check if Story is finished now (with last Paragraph marked Finished)
				count, err := CountFinishedParagraphsTx(tx, paragraph.Story)
				if err != nil {
					s.logger.Error(err)
					// tx already rolledback in countFinishedParagraphsTx
					return err
				}

				if count >= 7 {
					// Check if Story is marked finished already or not, update isFinished.
					story, err := GetStoryTx(tx, paragraph.Story)
					if err != nil {
						s.logger.Error(err)
						return err
					}

					if !story.IsFinished {
						// Mark Story as Finished.
						story.IsFinished = true
						if err := UpdateStoryTx(tx, *story); err != nil {
							s.logger.Error(err)
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
		s.logger.Error("Error executing transaction:" + err.Error())
		return errors.New("Errors executing transaction...")
	}

	return nil
}

// Get Sentence By ID (Transaction)
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

// Get All Sentences for a Paragraph ID (Transaction)
func GetParagraphSentencesTx(tx *sql.Tx, paragraphId int32) ([]string, error) {
	var sentences []string

	query := sq.Select("content").From("sentences").Where(sq.Eq{"paragraph": paragraphId})
	rows, err := query.RunWith(tx).Query()

	if err != nil {
		tx.Rollback()
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

// Update A Sentence (Transaction)
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
		tx.Rollback()
		return errors.New("Error Generating Sentence Update Query:" + err.Error())
	}

	if _, err = tx.Exec(query, args...); err != nil {
		tx.Rollback()
		return errors.New("Error Updating Sentence In DB:" + err.Error())
	}
	return nil
}
