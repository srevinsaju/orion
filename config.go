package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
)

/* CreateConfig helps to interactively create a configuration
from user input. */
func CreateConfig() {
	cfg := Config{}
	cfg.Channels = map[TelegramChannel]ChannelConfig{}

	var inputBuf string

	// GetChannel Telegram Api Token from @BotFather
	fmt.Print("Enter Telegram API Token: ")
	_, err := fmt.Scanln(&inputBuf)
	if err != nil {
		logger.Fatal(err)
		return
	}
	fmt.Println("")
	cfg.TelegramApiToken = inputBuf

	// GetChannel DiscordApiToken from Discord Application portal
	fmt.Print("Enter Discord API Token: ")
	_, err = fmt.Scanln(&inputBuf)
	if err != nil {
		logger.Fatal(err)
		return
	}
	fmt.Println("")

	for true {
		fmt.Println("Enter Telegram ChanID")
		fmt.Println("a comma, for example: -432232xx")
		_, err = fmt.Scanln(&inputBuf)

		if inputBuf == "EXIT" || err != nil {
			break
		}

		// do some checks on the telegram Channel ID
		telegramChanId, err := strconv.Atoi(inputBuf)
		if err != nil {
			logger.Warnf("%s is not a valid telegram channel id", inputBuf)
			continue
		}

		telegramChanIdTyped := TelegramChannel(telegramChanId)

		cfg.Channels[telegramChanIdTyped] = ChannelConfig{Reminder: map[string]Reminders{}}

	}

	outputBytes, err := json.MarshalIndent(cfg, "", "\t")
	err = ioutil.WriteFile("gofer.json", outputBytes, 0644)
	if err != nil {
		logger.Fatal(err)
		return
	}

}

/* ConfigFromFile creates a Config object from a JSON configuration file */
func ConfigFromFile(filepath string) (*Config, error) {
	rawData, err := ioutil.ReadFile(filepath)
	var cfg *Config
	err = json.Unmarshal(rawData, &cfg)
	if err != nil {
		logger.Fatal(err)
		return nil, err
	}
	return cfg, nil
}
