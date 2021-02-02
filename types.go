package main

import (
	"encoding/json"
	"io/ioutil"
)

type TelegramChannel int64


type Reminders struct {
	When string `json:"when"`
	What string `json:"what"`
}

type ChannelConfig struct {
	Reminder map[string]Reminders `json:"reminder,omitempty"`
}


type Config struct {
	TimeZone         string                            `json:"time_zone"`
	Channels         map[TelegramChannel]ChannelConfig `json:"channels"`
	TelegramApiToken string                            `json:"telegramApiToken"`
	configPath       string
}



func (config *Config) Write() {
	raw, err := json.Marshal(config)
	if err != nil {
		logger.Fatal("Failed to marshall JSON, %s", err)
		return
	}

	err = ioutil.WriteFile(config.configPath, raw, 0644)
	if err != nil {
		logger.Fatal("Failed to write configuration file to %s, %s", config.configPath, err)
	}
}





