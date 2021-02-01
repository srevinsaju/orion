package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

/* OnPinMessageHandler retrieves the message which was quoted and then pins it */
func OnPinMessageHandler(telegramBot *tgbotapi.BotAPI, update tgbotapi.Update, _ string){
	if update.Message.ReplyToMessage == nil {
		// the message has no reply
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Soz, donno which message to pin. 😎")
		msg.ReplyToMessageID = update.Message.MessageID
		_, err := telegramBot.Send(msg)
		if err != nil {
			logger.Warnf("Couldn't send message without reply to message, %s", err)

		}
		return
	}
	_, err := telegramBot.PinChatMessage(tgbotapi.PinChatMessageConfig{
		ChatID:              update.Message.ReplyToMessage.Chat.ID,
		MessageID:           update.Message.ReplyToMessage.MessageID,
		DisableNotification: false,
	})
	if err != nil {
		logger.Warnf("Failed to pin message, %s", err)
	}
}

/* OnMeMessageHandler handles messages which starts with /me and converts them to a familiar IRC-like statuses */
func OnMeMessageHandler(telegramBot *tgbotapi.BotAPI, update tgbotapi.Update, arguments string){
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		fmt.Sprintf("_* %s %s_", update.Message.From.FirstName, arguments))
	msg.ParseMode = "markdown"
	_, err := telegramBot.Send(msg)
	if err != nil {
		logger.Warnf("Couldn't send message without reply to message, %s", err)
	}

}


/* OnMessageNotCommandMatchHandler matches those messages which have no associated commands with them */
func OnMessageNotCommandMatchHandler(telegramBot *tgbotapi.BotAPI, update tgbotapi.Update, _ string) {
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID, "You would have to ask @sugaroidbot bro to answer this.")
	msg.ReplyToMessageID = update.Message.MessageID
	_, err := telegramBot.Send(msg)
	if err != nil {
		logger.Warnf("Couldn't send message without reply to message, %s", err)

	}
}

/* TelegramOnMessageHandler function scans every message and selects only those messages which can be processed */
func TelegramOnMessageHandler(telegramBot *tgbotapi.BotAPI, update tgbotapi.Update, config Config) {
	// first check if the message is from a valid registered channel
	if _, ok := config.Channels[TelegramChannel(update.Message.Chat.ID)]; !ok {
		logger.Infof("[TelegramBot] Received an event from unrecognized channel")
		return
	}

	// return if the message is empty
	if update.Message.Text == "" {
		return
	}

	command, arguments, err := GetCommandArgumentFromMessage(telegramBot, update)
	if err != nil {
		return
	}

	var handler func (telegramBot *tgbotapi.BotAPI, update tgbotapi.Update, arguments string)
	switch command {
	case "pin":
		handler = OnPinMessageHandler
	case "me":
		handler = OnMeMessageHandler
	default:
		handler = OnMessageNotCommandMatchHandler
	}

	// call the handler
	logger.Debugf("Calling the handler for command[%s] with arguments [%s]", command, arguments)
	handler(telegramBot, update, arguments)
}

/* TelegramEventHandler function is a long running function which scans all the incoming events */
func TelegramEventHandler(telegramBot *tgbotapi.BotAPI, config Config) {

	logger.Infof("[TelegramBot] Authorized on account %s", telegramBot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := telegramBot.GetUpdatesChan(u)
	if err != nil {
		logger.Fatal("Failed to get updates channel")
		return
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		// call the handler
		TelegramOnMessageHandler(telegramBot, update, config)
	}
}
