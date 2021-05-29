package main

import (
	"encoding/json"
	"io/ioutil"
)

type TelegramChannel int64

type Reminders struct {
	When       string `json:"when"`
	What       string `json:"what"`
	instanceId int
}

type ChannelConfig struct {
	Reminder            map[string]Reminders `json:"reminder,omitempty"`
	TimeZone            string               `json:"time_zone"`
	DiscordHaikuWebhook string               `json:"discord_haiku_webhook,omitempty"`
}

type Config struct {
	Version          int                               `json:"version,omitempty"`
	VerseSubscribers []int                             `json:"verse_subscribers,omitempty"`
	ChanAttr         map[TelegramChannel]TempChanAttr  `json:"chan_attr"`
	Channels         map[TelegramChannel]ChannelConfig `json:"channels"`
	TelegramApiToken string                            `json:"telegramApiToken"`
	GitHubApiToken   string                            `json:"github_api_token"`
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

type Quote struct {
	Text   string `json:"quote"`
	Author string `json:"author"`
	Title  string `json:"title"`
}

type Content struct {
	Quotes []Quote `json:"quotes"`
}

type QuoteOfTheDay struct {
	Contents Content `json:"contents"`
}

type VerseDetails struct {
	Text      string `json:"text"`
	Reference string `json:"reference"`
	Version   string `json:"version"`
}

type VerseMeta struct {
	Details VerseDetails `json:"details"`
}

type VOTD struct {
	Verse VerseMeta `json:"verse"`
}
