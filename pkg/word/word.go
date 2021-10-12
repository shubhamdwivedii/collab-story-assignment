package word

import (
	"errors"
	"fmt"

	. "github.com/shubhamdwivedii/collab-story-assignment/pkg/paragraph"
	. "github.com/shubhamdwivedii/collab-story-assignment/pkg/sentence"
	. "github.com/shubhamdwivedii/collab-story-assignment/pkg/story"
)

type WordStorage interface {
	GetUnfinishedStory() (*Story, error)
	AddStory() (int32, error)
	UpdateStoryTitle(storyId int32, word string) error

	GetUnfinishedParagraph(storyId int32) (*Paragraph, error)
	AddParagraph(storyId int32) (int32, error)

	GetUnfinishedSentence(paragraphId int32) (*Sentence, error)
	AddSentence(paragraphId int32, word string) (int32, error)
	UpdateSentence(sentenceId int32, word string) error
}

type WordService struct {
	storage WordStorage
}

func NewWordService(storage WordStorage) *WordService {
	wrdsrv := new(WordService)
	wrdsrv.storage = storage
	return wrdsrv
}

func (srv *WordService) AddWord(word string) error {
	if err := validateWord(word); err != nil {
		return err
	}

	// Find Unifinished Story
	story, err := srv.storage.GetUnfinishedStory()
	if err != nil {
		// No Unfinished Story, Create New Story
		storyId, err := srv.storage.AddStory()
		fmt.Println("Creating New Story")
		if err != nil { // New Story Cannot Be Created.
			return err
		}

		// New Story will have blank title, Add Word to title.
		if err := srv.storage.UpdateStoryTitle(storyId, word); err != nil {
			return err
		}
		return nil
	} else {
		fmt.Println("Story Found:", story)
		// // Unfinished Story Found, Check if story title is finished
		if !story.TitleAdded {
			// Add word to title instead.
			if err := srv.storage.UpdateStoryTitle(story.ID, word); err != nil {
				return err
			}
			return nil
		}

		// Story Title is Finished, Find Unfinished Paragraph
		paragraph, err := srv.storage.GetUnfinishedParagraph(story.ID)

		if err != nil {
			// No Unfinished Paragraph Found, Create New Paragraph
			paragraphId, err := srv.storage.AddParagraph(story.ID)
			fmt.Println("Creating New Paragraph")
			if err != nil {
				// New Paragraph Cannot Be Created.
				return err
			}

			// New Paragraph Created, Create a new Sentence.
			if _, err = srv.storage.AddSentence(paragraphId, word); err != nil {
				return err
			}
			// Word added to new Sentence.
			return nil
		} else {
			// Unfinished Paragraph Found, Find Unfinished Sentence.
			sentence, err := srv.storage.GetUnfinishedSentence(paragraph.ID)

			if err != nil {
				// No Unfinished Sentence Found, Create New Sentence.
				if _, err = srv.storage.AddSentence(paragraph.ID, word); err != nil {
					return err
				}
				// Word added to new Sentence.
				return nil
			} else {
				// Unfinished Sentence Found, add word to sentence
				if err := srv.storage.UpdateSentence(sentence.ID, word); err != nil {
					return err
				}
				// This will take care of updating both Paragraph and Story as finished if they are.
			}
		}
	}
	return nil
}

func validateWord(word string) error {
	if len(word) < 1 || len(word) > 240 {
		return errors.New("Invalid Word")
	}
	return nil
}

// Add mutex locking

// Add Tests

// Try to break further into modules

// break sql module into seperate files

// export one struct with inclusion.

/*
Add Word
	Find Unfinished Story
		N - Create New Story - Add Word to Title
		Y - Check if Title Finished
			N - Add Word To Title
			Y - Mark Title Finished - Find Unfinished Paragraph
				N - Create New Paragraph - New Sentence - Add Word
				Y - Find Unfinished Sentence
					N - Create New Sentence - Add Word
					Y - Add Word - Check if Sentence Full (15 Words)
						N - Finished.
						Y - Mark Sentence Finished - Check If Paragraph Full (10 Sentences)
							N - Finished.
							Y - Mark Paragraph Finished - Check If Story Full (7 Paragraphs)
								N - Finished.
								Y - Mark Story Finished.
*/
