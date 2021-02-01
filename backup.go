package main

type Message struct {
	Id        int    `json:"id"`
	Type      string `json:"type"`
	Action    string `json:"action,omitempty"`
	MessageId int    `json:"message_id,omitempty"`
}

type TelegramJSONBackup struct {
	Name     string    `json:"name"`
	Id       int64     `json:"id"`
	Messages []Message `json:"messages"`
}
