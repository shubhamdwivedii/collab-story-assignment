package sentence

type Sentence struct {
	ID         int32 `json:"id"`
	Paragraph  int32 `json:"paragraph"`
	IsFinished bool  `json:"is_finished"`
	Content    string
}

func (sentece *Sentence) Validate() error {
	return nil
}
