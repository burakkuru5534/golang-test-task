package model

type MessageData struct {
	ID       int64  `json:"id"`
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Message  string `json:"message"`
}
