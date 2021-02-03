package main

import (
	"encoding/json"
	"io/ioutil"
)

type TelegramChannel int64

type Reminders struct {
	When string `json:"when"`
	What string `json:"what"`
	instanceId int
}

type ChannelConfig struct {
	Reminder		map[string]Reminders `json:"reminder,omitempty"`
	TimeZone        string               `json:"time_zone"`
}

type Config struct {
	Channels         map[TelegramChannel]ChannelConfig `json:"channels"`
	TelegramApiToken string                            `json:"telegramApiToken"`
	configPath       string
}

func (config *Config) Write() {
	raw, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		logger.Fatal("Failed to marshall JSON, %s", err)
		return
	}

	err = ioutil.WriteFile(config.configPath, raw, 0644)
	if err != nil {
		logger.Fatal("Failed to write configuration file to %s, %s", config.configPath, err)
	}
}
