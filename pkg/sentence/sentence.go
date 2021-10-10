package sentence

type Sentence struct {
	ID        int32 `json:"id"`
	Paragraph int32 `json:"paragraph"`
	// CreatedAt time.Time `json:"created_at"`
	// UpdatedAt time.Time `json:"updated_at"`
	IsFinished bool `json:"is_finished"`
	Content    string
}
