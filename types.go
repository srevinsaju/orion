package main

type TelegramChannel int64


type Config struct {
	Channels         map[TelegramChannel]bool 			`json:"channels"`
	TelegramApiToken string                             `json:"telegramApiToken"`
}
