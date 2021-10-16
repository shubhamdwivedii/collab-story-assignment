package word

import (
	"errors"
	"fmt"
	"regexp"
	"sync"

	. "github.com/shubhamdwivedii/collab-story-assignment/pkg/paragraph"
	. "github.com/shubhamdwivedii/collab-story-assignment/pkg/sentence"
	. "github.com/shubhamdwivedii/collab-story-assignment/pkg/story"
)

type WordRequest struct { // add word request body.
	Word string `json:"word"`
}

type WordResponse struct {
	ID      int32  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"current_sentence"`
}

type WordError struct {
	Error string `json:"error"`
}

type WordStorage interface {
	GetUnfinishedStory() (*Story, error)
	AddStory() (int32, error)
	GetStory(storyId int32) (*Story, error)
	UpdateStoryTitle(storyId int32, word string) error

	GetUnfinishedParagraph(storyId int32) (*Paragraph, error)
	AddParagraph(storyId int32) (int32, error)

	GetUnfinishedSentence(paragraphId int32) (*Sentence, error)
	AddSentence(paragraphId int32, word string) (int32, error)
	GetSentence(sentenceId int32) (*Sentence, error)
	UpdateSentence(sentenceId int32, word string) error
}

type WordService struct {
	storage WordStorage
	sync.Mutex
}

func NewWordService(storage WordStorage) *WordService {
	wrdsrv := new(WordService)
	wrdsrv.storage = storage
	return wrdsrv
}

func (srv *WordService) AddWord(word string) (*WordResponse, error) {
	if err := validateWord(word); err != nil {
		return nil, err
	}

	defer srv.Unlock()
	// Adding two words concurrently might lead to inconsistency.
	srv.Lock()

	// Find Unifinished Story
	story, err := srv.storage.GetUnfinishedStory()
	if err != nil {
		// No Unfinished Story, Create New Story
		storyId, err := srv.storage.AddStory()
		fmt.Println("Creating New Story")
		if err != nil { // New Story Cannot Be Created.
			return nil, err
		}

		// New Story will have blank title, Add Word to title.
		if err := srv.storage.UpdateStoryTitle(storyId, word); err != nil {
			return nil, err
		}

		// get story by id
		story, err := srv.storage.GetStory(storyId)
		if err != nil {
			return nil, err
		}
		wrdRes := WordResponse{
			ID:      story.ID,
			Title:   story.Title,
			Content: "",
		}
		return &wrdRes, nil
	} else {
		fmt.Println("Story Found:", story)
		// // Unfinished Story Found, Check if story title is finished
		if !story.TitleAdded {
			// Add word to title instead.
			if err := srv.storage.UpdateStoryTitle(story.ID, word); err != nil {
				return nil, err
			}
			// get story by id
			// get story by id
			story, err := srv.storage.GetStory(story.ID)
			if err != nil {
				return nil, err
			}
			wrdRes := WordResponse{
				ID:      story.ID,
				Title:   story.Title,
				Content: "",
			}
			return &wrdRes, nil
		}

		// Story Title is Finished, Find Unfinished Paragraph
		paragraph, err := srv.storage.GetUnfinishedParagraph(story.ID)

		if err != nil {
			// No Unfinished Paragraph Found, Create New Paragraph
			paragraphId, err := srv.storage.AddParagraph(story.ID)
			fmt.Println("Creating New Paragraph")
			if err != nil {
				// New Paragraph Cannot Be Created.
				return nil, err
			}

			// New Paragraph Created, Create a new Sentence.
			if sentenceId, err := srv.storage.AddSentence(paragraphId, word); err != nil {
				return nil, err
			} else {
				// Word added to new Sentence.
				// get sentence by id
				sentence, err := srv.storage.GetSentence(sentenceId)
				if err != nil {
					return nil, err
				}
				wrdRes := WordResponse{
					ID:      story.ID,
					Title:   story.Title,
					Content: sentence.Content,
				}
				return &wrdRes, nil
			}
		} else {
			// Unfinished Paragraph Found, Find Unfinished Sentence.
			sentence, err := srv.storage.GetUnfinishedSentence(paragraph.ID)

			if err != nil {
				// No Unfinished Sentence Found, Create New Sentence.
				if sentenceId, err := srv.storage.AddSentence(paragraph.ID, word); err != nil {
					return nil, err
				} else {
					// Word added to new Sentence.
					// get sentence by id
					sentence, err := srv.storage.GetSentence(sentenceId)
					if err != nil {
						return nil, err
					}
					wrdRes := WordResponse{
						ID:      story.ID,
						Title:   story.Title,
						Content: sentence.Content,
					}
					return &wrdRes, nil
				}
			} else {
				// Unfinished Sentence Found, add word to sentence
				if err := srv.storage.UpdateSentence(sentence.ID, word); err != nil {
					return nil, err
				} else {
					// sentence updated
					sentence, err := srv.storage.GetSentence(sentence.ID)
					if err != nil {
						return nil, err
					}
					wrdRes := WordResponse{
						ID:      story.ID,
						Title:   story.Title,
						Content: sentence.Content,
					}
					return &wrdRes, nil
				}
				// This will take care of updating both Paragraph and Story as finished if they are.
			}
		}
	}
	// return nil
}

func validateWord(word string) error {
	if len(word) < 1 || len(word) > 240 {
		return errors.New("Invalid Word Length")
	} else {
		re, _ := regexp.Compile("^\\S+$")
		if re.MatchString(word) {
			// word contains no spaces
			return nil
		} else {
			return errors.New("Multiple Words Sent")
		}
	}
}

// Add mutex locking
// Add Tests

// Try to break further into modules

// break sql module into seperate files

// export one struct with inclusion.
