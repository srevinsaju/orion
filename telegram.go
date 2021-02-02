package main

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/robfig/cron"
	"strings"
	"time"
)

type MessageHandlerArgs struct {
	bot 	  *tgbotapi.BotAPI
	update 	  tgbotapi.Update
	arguments string
	config 	  *Config
	cronMgr   *cron.Cron
}


func getChannel(h MessageHandlerArgs) (*ChannelConfig, error) {
	val, ok := h.config.Channels[TelegramChannel(h.update.Message.Chat.ID)]
	if ! ok {
		return nil, errors.New("did not find channel id")
	}
	return &val, nil
}

/* OnPinMessageHandler retrieves the message which was quoted and then pins it */
func OnPinMessageHandler(h MessageHandlerArgs){
	if h.update.Message.ReplyToMessage == nil {
		// the message has no reply
		msg := tgbotapi.NewMessage(h.update.Message.Chat.ID, "Soz, donno which message to pin. 😎")
		msg.ReplyToMessageID = h.update.Message.MessageID
		_, err := h.bot.Send(msg)
		if err != nil {
			logger.Warnf("Couldn't send message without reply to message, %s", err)

		}
		return
	}
	_, err := h.bot.PinChatMessage(tgbotapi.PinChatMessageConfig{
		ChatID:              h.update.Message.ReplyToMessage.Chat.ID,
		MessageID:           h.update.Message.ReplyToMessage.MessageID,
		DisableNotification: false,
	})
	if err != nil {
		logger.Warnf("Failed to pin message, %s", err)
	}
}

/* OnScheduleMessageHandler handles messages which starts with /me and converts them to a familiar IRC-like statuses */
func OnScheduleMessageHandler(h MessageHandlerArgs){

	msg := tgbotapi.NewMessage(h.update.Message.Chat.ID,"")

	arguments := strings.Split(h.arguments, ",")
	if len(arguments) < 2 {
		msg.Text = "Usage: cron job, reminder text"
		_, err := h.bot.Send(msg)
		if err != nil {
			logger.Warnf("Couldn't send message without reply to message, %s", err)
		}
		return
	}
	cronJob := arguments[0]
	reminderMessage := arguments[1]

	// -->
	val, err := getChannel(h)
	if err != nil {
		logger.Warnf("Failed to getChannel channel configuration, %s", h.update.Message.Chat.ID)
		return
	}

	val.Reminder[cronJob] = Reminders{
		When: cronJob,
		What: reminderMessage,
	}
	logger.Info(val.Reminder[cronJob], h.config)
	h.config.Write()


	err = h.cronMgr.AddFunc(cronJob, func() {
		logger.Infof("Triggering cronjob set at %s in %s", cronJob, h.update.Message.Chat.ID)
		msgCron := tgbotapi.NewMessage(h.update.Message.Chat.ID, reminderMessage)
		_, err_ := h.bot.Send(msgCron)
		if err_ != nil {
			logger.Warnf("Couldn't send message without reply to message, %s", err)
		}
	})
	logger.Infof("Setting cron at %s, %s", cronJob, err)

	if err != nil {
		msg.Text = fmt.Sprintf("⏰ invalid cron. Can't set the cron job, %s", err)
	} else {
		msg.Text = fmt.Sprintf("⏰ Successfully set reminder for %s at %s", reminderMessage, cronJob )
	}


	_, err = h.bot.Send(msg)
	if err != nil {
		logger.Warnf("Couldn't send message without reply to message, %s", err)
	}

}


/* OnMeMessageHandler handles messages which starts with /me and converts them to a familiar IRC-like statuses */
func OnMeMessageHandler(h MessageHandlerArgs){
	msg := tgbotapi.NewMessage(
		h.update.Message.Chat.ID,
		fmt.Sprintf("_* %s %s_", h.update.Message.From.FirstName, h.arguments))
	msg.ParseMode = "markdown"
	_, err := h.bot.Send(msg)
	if err != nil {
		logger.Warnf("Couldn't send message without reply to message, %s", err)
	}

}


/* OnMessageNotCommandMatchHandler matches those messages which have no associated commands with them */
func OnMessageNotCommandMatchHandler(h MessageHandlerArgs) {
	msg := tgbotapi.NewMessage(
		h.update.Message.Chat.ID, "You would have to ask @sugaroidbot bro to answer this.")
	msg.ReplyToMessageID = h.update.Message.MessageID
	_, err := h.bot.Send(msg)
	if err != nil {
		logger.Warnf("Couldn't send message without reply to message, %s", err)

	}
}

/* TelegramOnMessageHandler function scans every message and selects only those messages which can be processed */
func TelegramOnMessageHandler(h MessageHandlerArgs) {
	// first check if the message is from a valid registered channel
	if _, ok := h.config.Channels[TelegramChannel(h.update.Message.Chat.ID)]; !ok {
		logger.Infof("[TelegramBot] Received an event from unrecognized channel")
		return
	}

	// return if the message is empty
	if h.update.Message.Text == "" {
		return
	}

	command, arguments, err := GetCommandArgumentFromMessage(h.bot, h.update)
	if err != nil {
		return
	}

	var handler func (h MessageHandlerArgs)
	switch command {
	case "pin":
		handler = OnPinMessageHandler
	case "me":
		handler = OnMeMessageHandler
	case "schedule":
		handler = OnScheduleMessageHandler
	default:
		handler = OnMessageNotCommandMatchHandler
	}

	// call the handler
	logger.Debugf("Calling the handler for command[%s] with arguments [%s]", command, arguments)

	h.arguments = arguments
	handler(h)
}

/* TelegramEventHandler function is a long running function which scans all the incoming events */
func TelegramEventHandler(telegramBot *tgbotapi.BotAPI, config *Config) {

	logger.Infof("[TelegramBot] Authorized on account %s", telegramBot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := telegramBot.GetUpdatesChan(u)
	if err != nil {
		logger.Fatal("Failed to getChannel updates channel")
		return
	}

	// create the scheduler instance
	location, err := time.LoadLocation(config.TimeZone)
	if err != nil {
		location = time.UTC
	}
	c := cron.NewWithLocation(location)


	for chanId, chanInstance := range config.Channels {
		chanIdInt := int64(chanId)
		for k, v := range chanInstance.Reminder {

			err = c.AddFunc(k, func() {
				if err != nil {
					return
				}
				logger.Infof("Triggering cronjob set at %s in %s", k, chanId)
				msgCron := tgbotapi.NewMessage(chanIdInt, v.What)
				_, err_ := telegramBot.Send(msgCron)
				if err_ != nil {
					logger.Warnf("Couldn't send message without reply to message, %s", err)
				}
			})
			logger.Infof("Setting cron at %s, %s", k, err)
		}
	}

	go c.Start()

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		handlerArgs := &MessageHandlerArgs{
			bot:       telegramBot,
			update:    update,
			arguments: "",
			config:    config,
			cronMgr:   c,
		}

		// call the handler
		TelegramOnMessageHandler(*handlerArgs)
	}
}
