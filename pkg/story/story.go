package story

import "time"

type Story struct {
	ID         int32     `json:"id"`
	Title      string    `json:"title"`
	TitleAdded bool      `json:"title_added"`
	IsFinished bool      `json:"is_finished"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
