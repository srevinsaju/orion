package main

import (
	"context"
	"errors"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/go-github/github"
	"github.com/robfig/cron/v3"
)

type MessageHandlerArgs struct {
	bot       *tgbotapi.BotAPI
	update    tgbotapi.Update
	arguments string
	config    *Config
	cronMgr   *cron.Cron
	gh        *github.Client
	ghCtx     context.Context
}

type TempChanAttr struct {
	Pluses                int
	MessageCount          int
	LastMessageCountAsked int
}

var TempChanAttrs = map[TelegramChannel]*TempChanAttr{}

/* GetChannel gets the channel from the channel configuration. */
func GetChannel(h MessageHandlerArgs) (*ChannelConfig, error) {
	val, ok := h.config.Channels[TelegramChannel(h.update.Message.Chat.ID)]
	if !ok {
		return nil, errors.New("did not find channel id")
	}
	return &val, nil
}

/* TelegramOnMessageHandler function scans every message and selects only those messages which can be processed */
func TelegramOnMessageHandler(h MessageHandlerArgs) {
	// first check if the message is from a valid registered channel
	_, err := GetChannel(h)
	if err != nil {
		logger.Infof("[TelegramBot] Received an event from unrecognized channel")
		return
	}

	// return if the message is empty
	if h.update.Message.Text == "" {
		return
	}

	// add the number of pluses to the Pluses variable
	go func() {
		TempChanAttrs[TelegramChannel(h.update.Message.Chat.ID)].MessageCount += 1
		count := strings.Count(h.update.Message.Text, "++")
		TempChanAttrs[TelegramChannel(h.update.Message.Chat.ID)].Pluses += count
		if count > 0 {
			logger.Infof("Pluses added %d", count)
			logger.Infof("Number of pluses now, %d", TempChanAttrs[TelegramChannel(h.update.Message.Chat.ID)].Pluses)
		}
	}()

	go OnHaikuMessageHandler(h)

	// start another go routine for handling messages with sed
	messageTrimmedToLower := strings.Trim(strings.ToLower(h.update.Message.Text), " ")
	if strings.Contains(messageTrimmedToLower, "sed") {
		go OnSedMessageHandler(h)
	}

	if strings.Contains(messageTrimmedToLower, "😌😌😌") {
		go OnRelievedMessageHandler(h)
	}

	if strings.Contains(messageTrimmedToLower, "😔😔😔") {
		go OnSedMessageHandler(h)
	}

	if strings.Contains(messageTrimmedToLower, "🌞🌞🌞") {
		go OnSunnyMessageHandler(h)
	}

	if strings.Contains(messageTrimmedToLower, "twitter.com") {
		go OnTwitterMessageHandler(h)
	}

	// get the command and arguments
	command, arguments, err := GetCommandArgumentFromMessage(h.bot, h.update)
	if err != nil {
		return
	}

	// create a handler variable which can be later run as goroutine
	var handler func(h MessageHandlerArgs)
	switch command {
	case "pin":
		handler = OnPinMessageHandler
	case "me":
		handler = OnMeMessageHandler
	case "schedule":
		handler = OnScheduleMessageHandler
	case "unschedule":
		handler = OnUnScheduleMessageHandler
	case "listscheduled":
		handler = OnListScheduleMessageHandler
	case "plus":
		handler = OnPlusesMessageHandler
	case "tex":
		handler = OnLatexMessageHandler
	case "count":
		handler = OnCountMessageHandler
	case "report":
		handler = OnReportMessageHandler
	case "id":
		handler = OnIdMessageHandler
	case "8ball":
		handler = On8BallMessageHandler
	default:
		handler = OnMessageNotCommandMatchHandler
	}

	// call the handler
	logger.Debugf("Calling the handler for command[%s] with arguments [%s]", command, arguments)

	h.arguments = arguments
	go handler(h)
}

/* TelegramEventHandler function is a long running function which scans all the incoming events */
func TelegramEventHandler(
	telegramBot *tgbotapi.BotAPI, ghClient *github.Client, ghCtx context.Context, config *Config) {

	logger.Infof("[TelegramBot] Authorized on account %s", telegramBot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// check if the version has changed

	if config.Version != Version {
		handlerArgs := &MessageHandlerArgs{
			bot:       telegramBot,
			arguments: "",
			config:    config,
		}
		OnVersionChangedHandler(*handlerArgs)
	}

	updates, err := telegramBot.GetUpdatesChan(u)
	if err != nil {
		logger.Fatal("Failed to GetChannel updates channel")
		return
	}

	// initialize the message counters
	for k := range config.Channels {
		TempChanAttrs[k] = &TempChanAttr{
			Pluses:       config.ChanAttr[k].Pluses,
			MessageCount: config.ChanAttr[k].MessageCount,
		}
	}

	c := cron.New()
	// set the cron jobs
	ScheduleCronFromConfig(config, telegramBot, c)
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
			gh:        ghClient,
			ghCtx:     ghCtx,
		}

		// call the handler
		TelegramOnMessageHandler(*handlerArgs)
	}

}
