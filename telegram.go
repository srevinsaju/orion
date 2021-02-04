package main

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/robfig/cron/v3"
	"strings"
)

type MessageHandlerArgs struct {
	bot       *tgbotapi.BotAPI
	update    tgbotapi.Update
	arguments string
	config    *Config
	cronMgr   *cron.Cron
}

type TempChanAttr struct {
	Pluses       int
	MessageCount int
}

var TempChanAttrs = map[TelegramChannel]*TempChanAttr{}

func getChannel(h MessageHandlerArgs) (*ChannelConfig, error) {
	val, ok := h.config.Channels[TelegramChannel(h.update.Message.Chat.ID)]
	if !ok {
		return nil, errors.New("did not find channel id")
	}
	return &val, nil
}

/* OnPinMessageHandler retrieves the message which was quoted and then pins it */
func OnPinMessageHandler(h MessageHandlerArgs) {
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

/* OnScheduleMessageHandler handles messages which aim at scheduling using Cron Syntax */
func OnScheduleMessageHandler(h MessageHandlerArgs) {

	msg := tgbotapi.NewMessage(h.update.Message.Chat.ID, "")

	arguments := strings.Split(h.arguments, "@")
	if len(arguments) < 2 {
		msg.Text = "Usage: reminder text @ cron_expression"
		_, err := h.bot.Send(msg)
		if err != nil {
			logger.Warnf("Couldn't send message without reply to message, %s", err)
		}
		return
	}
	reminderMessage := arguments[0]
	cronJob := arguments[1]
	cronJob = strings.Trim(cronJob, " ")

	// -->
	val, err := getChannel(h)
	if err != nil {
		logger.Warnf("Failed to getChannel channel configuration, %s", h.update.Message.Chat.ID)
		return
	}

	instanceId, err := h.cronMgr.AddFunc(
		fmt.Sprintf("TZ=%s %s", val.TimeZone, cronJob),
		func() {
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
		msg.Text = fmt.Sprintf("⏰ Successfully set reminder for %s at %s", reminderMessage, cronJob)
		val.Reminder[cronJob] = Reminders{
			When:       cronJob,
			What:       reminderMessage,
			instanceId: int(instanceId),
		}
		logger.Info(val.Reminder[cronJob], h.config)
		h.config.Write()
	}

	_, err = h.bot.Send(msg)
	if err != nil {
		logger.Warnf("Couldn't send message without reply to message, %s", err)
	}

}

/* OnUnScheduleMessageHandler handles messages which aim at unscheduling using Cron Syntax */
func OnUnScheduleMessageHandler(h MessageHandlerArgs) {

	msg := tgbotapi.NewMessage(h.update.Message.Chat.ID, "")

	cronJob := strings.Trim(h.arguments, " ")

	// -->
	val, err := getChannel(h)
	if err != nil {
		logger.Warnf("Failed to getChannel channel configuration, %s", h.update.Message.Chat.ID)
		return
	}

	job, ok := val.Reminder[cronJob]
	if !ok {
		message := fmt.Sprintf("Failed to find cron job '<code>%s</code>' in the listing", cronJob)
		logger.Warn(message)
		msg.Text = message
	} else {
		logger.Debug("Deleting cronJob from local data")
		delete(val.Reminder, cronJob)
		logger.Debug("Updating configuration")
		h.config.Write()
		logger.Infof("Removing InstanceId[%s]", job.instanceId)
		h.cronMgr.Remove(cron.EntryID(job.instanceId))

		msg.Text = fmt.Sprintf("Removed reminder for <code>%s</code>. Updating cron jobs with new config", cronJob)
	}

	// set the parse mode to html
	msg.ParseMode = "html"
	_, err = h.bot.Send(msg)
	if err != nil {
		logger.Warnf("Couldn't send message without reply to message, %s", err)
	}

	msgSuccess := tgbotapi.NewMessage(h.update.Message.Chat.ID, "CronJobs reloaded 🚀")
	_, err = h.bot.Send(msgSuccess)
	if err != nil {
		logger.Warnf("Couldn't send message without reply to message, %s", err)
	}
}

/* OnListScheduleMessageHandler handles messages which aim at list scheduling using Cron Syntax */
func OnListScheduleMessageHandler(h MessageHandlerArgs) {

	msg := tgbotapi.NewMessage(h.update.Message.Chat.ID, "")

	// -->
	val, err := getChannel(h)
	if err != nil {
		logger.Warnf("Failed to getChannel channel configuration, %s", h.update.Message.Chat.ID)
		return
	}

	var scheduledCronJobsMessages []string
	for _, v := range val.Reminder {
		scheduledCronJobsMessages =
			append(scheduledCronJobsMessages, fmt.Sprintf("✨ <code>%s</code> 🢂 <i>%s</i>", v.When, v.What))
	}
	messages := strings.Join(scheduledCronJobsMessages, string('\n'))
	messages = "<b>🚨 Scheduled Reminders</b> \n\n" + messages

	msg.Text = messages
	msg.ParseMode = "html"
	_, err = h.bot.Send(msg)
	if err != nil {
		logger.Warnf("Couldn't send message without reply to message, %s", err)
	}
}

/* OnPlusesMessageHandler handles messages which starts with /plus and returns the total number of pluses */
func OnPlusesMessageHandler(h MessageHandlerArgs) {
	msg := tgbotapi.NewMessage(
		h.update.Message.Chat.ID,
		fmt.Sprintf("Pluses so far: <b>%d</b>", TempChanAttrs[TelegramChannel(h.update.Message.Chat.ID)].Pluses))
	msg.ParseMode = "html"
	_, err := h.bot.Send(msg)
	if err != nil {
		logger.Warnf("Couldn't send message without reply to message, %s", err)
	}

}

/* OnCountMessageHandler handles messages which starts with /count and returns the total number of messages */
func OnCountMessageHandler(h MessageHandlerArgs) {
	msg := tgbotapi.NewMessage(
		h.update.Message.Chat.ID,
		fmt.Sprintf("Messages so far: <b>%d</b>",
			TempChanAttrs[TelegramChannel(h.update.Message.Chat.ID)].MessageCount))
	msg.ParseMode = "html"
	_, err := h.bot.Send(msg)
	if err != nil {
		logger.Warnf("Couldn't send message without reply to message, %s", err)
	}

}

/* OnMeMessageHandler handles messages which starts with /me and converts them to a familiar IRC-like statuses */
func OnMeMessageHandler(h MessageHandlerArgs) {
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
	_, err := getChannel(h)
	if err != nil {
		logger.Infof("[TelegramBot] Received an event from unrecognized channel")
		return
	}

	// return if the message is empty
	if h.update.Message.Text == "" {
		return
	}

	TempChanAttrs[TelegramChannel(h.update.Message.Chat.ID)].MessageCount += 1
	count := strings.Count(h.update.Message.Text, "++")
	TempChanAttrs[TelegramChannel(h.update.Message.Chat.ID)].Pluses += count

	if count > 0 {
		logger.Infof("Pluses added %d", count)
		logger.Infof("Number of pluses now, %d", TempChanAttrs[TelegramChannel(h.update.Message.Chat.ID)].Pluses)
	}

	command, arguments, err := GetCommandArgumentFromMessage(h.bot, h.update)
	if err != nil {
		return
	}

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
	case "count":
		handler = OnCountMessageHandler
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
		}

		// call the handler
		TelegramOnMessageHandler(*handlerArgs)
	}

}
