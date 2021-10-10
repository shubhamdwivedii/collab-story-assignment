package paragraph

type Paragraph struct {
	ID         int32 `json:"id"`
	Story      int32 `json:"story"`
	IsFinished bool  `json:"is_finished"`
}
