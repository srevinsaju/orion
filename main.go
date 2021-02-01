package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/withmandala/go-log"
	"os"
)

var logger = log.New(os.Stdout)

func main() {
	// getChannel last command
	command := os.Args[len(os.Args)-1]
	if command == "create" {
		CreateConfig()
		logger.Info("Config file written. Call `gofer` like")
		logger.Infof("~ $ orion path/to/config.json")
		return
	}

	if command == "orion" {
		// the user has not provided any commands along with the executable name
		// so, we should show the usage
		logger.Info("orion : yet another telegram pin bot")
		logger.Info("")
		logger.Info("To load an existing configuration: ")
		logger.Info("  $ orion path/to/config.json")
		logger.Info("")
		logger.Info("To create a new configuration in current directory:")
		logger.Info("  $ orion create")
		return

	}

	if _, err := os.Stat(command); os.IsNotExist(err) {
		logger.Fatal("The specified path does not exist")
	}

	goferCfgFile := command

	cfg, err := ConfigFromFile(goferCfgFile)
	if err != nil {
		logger.Fatal(err)
	}
	// set the config path to the command name
	cfg.configPath = command

	telegramBotToken := cfg.TelegramApiToken
	// create the telegram bot
	telegramBot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		logger.Fatal(err)
	}

	telegramBot.Debug = false


	TelegramEventHandler(telegramBot, cfg)

}
