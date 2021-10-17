package word

import (
	"errors"
	"regexp"
	"sync"

	. "github.com/shubhamdwivedii/collab-story/pkg/paragraph"
	. "github.com/shubhamdwivedii/collab-story/pkg/sentence"
	. "github.com/shubhamdwivedii/collab-story/pkg/story"
	log "github.com/sirupsen/logrus"
)

type WordRequest struct {
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
	logger  *log.Logger
	sync.Mutex
}

func NewWordService(storage WordStorage, logger *log.Logger) *WordService {
	wrdsrv := new(WordService)
	wrdsrv.storage = storage
	wrdsrv.logger = logger
	return wrdsrv
}

func ValidateWord(word string) error {
	if len(word) < 1 || len(word) > 16 {
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

// This will add a Word to a Story/Paragraph/Sentence in Storage
func (srv *WordService) AddWord(word string) (*WordResponse, error) {
	if err := ValidateWord(word); err != nil {
		srv.logger.Error("Word is invalid.")
		return nil, err
	}

	defer srv.Unlock()
	// Adding two words concurrently might lead to inconsistency
	srv.Lock()

	// Find Unfinished Story
	story, err := srv.storage.GetUnfinishedStory()
	if err != nil {
		// No Unfinished Story, Create New Story
		srv.logger.Info("Unfinished Story Not Found, Creating New Story...")
		storyId, err := srv.storage.AddStory()
		if err != nil {
			srv.logger.Error("Could Not Create New Story.")
			return nil, err
		}

		// New Story will have blank title, Add Word to title.
		if err := srv.storage.UpdateStoryTitle(storyId, word); err != nil {
			srv.logger.Error("Could Not Update Story Title.")
			return nil, err
		} else {
			// Word Added To Story Title, Get Story By ID
			story, err := srv.storage.GetStory(storyId)
			if err != nil {
				srv.logger.Error("Unexpected Error: Story Missing:" + err.Error())
				return nil, err
			}
			wrdRes := WordResponse{
				ID:      story.ID,
				Title:   story.Title,
				Content: "",
			}
			return &wrdRes, nil
		}
	} else {
		// Unfinished Story Found, Check if Story's Title is Finished.
		if !story.TitleAdded {
			// Add Word to Title instead.
			if err := srv.storage.UpdateStoryTitle(story.ID, word); err != nil {
				srv.logger.Error("Could Not Update Story Title.")
				return nil, err
			} else {
				// Word Added To Story Title, Get Story By ID
				story, err := srv.storage.GetStory(story.ID)
				if err != nil {
					srv.logger.Error("Unexpected Error: Story Missing:" + err.Error())
					return nil, err
				}
				wrdRes := WordResponse{
					ID:      story.ID,
					Title:   story.Title,
					Content: "",
				}
				return &wrdRes, nil
			}
		}

		// Story Title is already Finished, Find Unfinished Paragraph
		paragraph, err := srv.storage.GetUnfinishedParagraph(story.ID)

		if err != nil {
			// No Unfinished Paragraph Found, Create New Paragraph
			srv.logger.Info("Unfinished Paragraph Not Found, Creating New Paragraph...")
			paragraphId, err := srv.storage.AddParagraph(story.ID)
			if err != nil {
				// New Paragraph Cannot Be Created.
				srv.logger.Error("Could Not Create New Paragraph.")
				return nil, err
			}

			// New Paragraph Created, Create a new Sentence (with Word).
			if sentenceId, err := srv.storage.AddSentence(paragraphId, word); err != nil {
				srv.logger.Error("Could Not Add New Sentence")
				return nil, err
			} else {
				// Word Added To Sentence, Get Sentence By ID
				sentence, err := srv.storage.GetSentence(sentenceId)
				if err != nil {
					srv.logger.Error("Unexpected Error: Sentence Missing:" + err.Error())
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
			// Unfinished Paragraph Found. Find Unfinished Sentence.
			sentence, err := srv.storage.GetUnfinishedSentence(paragraph.ID)
			if err != nil {
				// No Unfinished Sentence Found, Create New Sentence (with Word).
				srv.logger.Info("Unfinished Sentence Not Found, Creating New Sentence...")
				if sentenceId, err := srv.storage.AddSentence(paragraph.ID, word); err != nil {
					srv.logger.Error("Could Not Add New Sentence")
					return nil, err
				} else {
					// Word Added To New Sentence. Get Sentence By ID
					sentence, err := srv.storage.GetSentence(sentenceId)
					if err != nil {
						srv.logger.Error("Unexpected Error: Sentence Missing:" + err.Error())
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
				// Unfinished Sentence Found, Add Word to Sentence
				if err := srv.storage.UpdateSentence(sentence.ID, word); err != nil {
					srv.logger.Error("Could Not Update Sentence.")
					return nil, err
				} else {
					// Word Added To Existing Sentence. Get Sentence By ID
					sentence, err := srv.storage.GetSentence(sentence.ID)
					if err != nil {
						srv.logger.Error("Unexpected Error: Sentence Missing:" + err.Error())
						return nil, err
					}
					wrdRes := WordResponse{
						ID:      story.ID,
						Title:   story.Title,
						Content: sentence.Content,
					}
					return &wrdRes, nil
				}
			}
			// NOTE: Updating a Sentence will take care of updating both Paragraph and Story as finished if they are.
		}
	}
}

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
