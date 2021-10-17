package story

import (
	"time"

	log "github.com/sirupsen/logrus"
)

type Story struct {
	ID         int32     `json:"id"`
	Title      string    `json:"title"`
	TitleAdded bool      `json:"title_added"`
	IsFinished bool      `json:"is_finished"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type StoryBrief struct {
	ID        int32     `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StoriesResponse struct {
	Limit   int32        `json:"limit"`
	Offset  int32        `json:"offset"`
	Count   int32        `json:"count"`
	Results []StoryBrief `json:"results"`
}

type StoryResponse struct {
	ID         int32            `json:"id"`
	Title      string           `json:"title"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
	Paragraphs []ParagraphBrief `json:"paragraphs"`
}

type ParagraphBrief struct {
	ID        int32    `json:"id"`
	Sentences []string `json:"sentences"`
}

type StoryStorage interface {
	GetAllStories(limit int32, offset int32) (*StoriesResponse, error)
	GetStoryDetail(storyId int32) (*StoryResponse, error)
}

type StoryService struct {
	storage StoryStorage
	logger  *log.Logger
	// sync.Mutex // not needed for read only operation (tx implemented in sql)
}

func NewStoryService(storage StoryStorage, logger *log.Logger) *StoryService {
	strsrv := new(StoryService)
	strsrv.storage = storage
	strsrv.logger = logger
	return strsrv
}

func (srv *StoryService) GetAllStories(limit int32, offset int32) (*StoriesResponse, error) {
	// Add Some Metrics Here ? Or Some Business Logic
	return srv.storage.GetAllStories(limit, offset)
}

func (srv *StoryService) GetStoryDetail(storyId int32) (*StoryResponse, error) {
	// Add Some Metrics Here ? Or Some Business Logic
	return srv.storage.GetStoryDetail(storyId)
}
