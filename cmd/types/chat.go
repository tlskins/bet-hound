package types

import (
	"time"
)

type Chatroom struct {
	Name      string
	Messages  []Message
	Observers map[string]struct {
		Username string
		Message  chan *Message
	}
}

type Message struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	CreatedBy string    `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
}
