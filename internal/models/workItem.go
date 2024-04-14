package models

type WorkItem struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
	Seq int    `json:"seq,omitempty"`
}
