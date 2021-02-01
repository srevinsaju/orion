package main

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)


/* GetCommandArgumentFromMessage splits the message into the command and the arguments which are to be
 passed to it */
func GetCommandArgumentFromMessage(telegramBot *tgbotapi.BotAPI, update tgbotapi.Update) (string, string, error) {
	var command string
	var arguments string
	if update.Message.IsCommand() {
		command = update.Message.Command()
		arguments = update.Message.CommandArguments()
	} else {

		words := strings.Split(update.Message.Text, string(' '))
		command = words[0]
		if command != fmt.Sprintf("@%s", telegramBot.Self.UserName) {
			// the command does not mention the bot
			return "", "", errors.New("the message doesnt start with telegramBot.Self.UserName")
		}

		command = words[1]
		arguments = strings.Join(words[2:], string(' '))
	}
	return command, arguments, nil
}