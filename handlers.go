package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/go-github/github"
	"github.com/robfig/cron/v3"
	"math/rand"
	"strings"
	"time"
)

var RandomSource = rand.NewSource(time.Now().Unix())
var Random = rand.New(RandomSource)

var SedResponses = []string{
	"So sed.",
	"sed sed.",
	"sed indeed.",
	"sed.",
	"sed tbh.",
	"sed, but idk.",
	"sed, but ok.",
	"sed sed sed",
	"super sed.",
}

/* OnReportMessageHandler retrieves the message which was quoted and then pins it */
func OnReportMessageHandler(h MessageHandlerArgs) {
	if h.update.Message.ReplyToMessage == nil {
		// the message has no reply
		msg := tgbotapi.NewMessage(h.update.Message.Chat.ID, "Soz, donno which message to report. Sed. 😎")
		msg.ReplyToMessageID = h.update.Message.MessageID
		_, err := h.bot.Send(msg)
		if err != nil {
			logger.Warnf("Couldn't send message without reply to message, %s", err)

		}
		return
	}
	if strings.Trim(h.update.Message.CommandArguments(), " ") == "" {
		// no message along with the command
		SendMessage("Usage: /report <the question you asked to sugaroid>", h)
		return
	}

	body := fmt.Sprintf(`
	The following answer was a given by sugaroid, and is considered as a bug. Please 
	manually review this issue, and close it if it looks unnecessary.

	user: %s 
    sugaroid: %s`,
		h.update.Message.CommandArguments(), h.update.Message.ReplyToMessage.Text,
	)
	title := fmt.Sprintf("[+orion] Bug '%s'", h.update.Message.CommandArguments())
	labels := []string{"orion"}

	issue, _, err := h.gh.Issues.Create(h.ghCtx, "sugaroidbot", "sugaroid", &github.IssueRequest{
		Title:  &title,
		Body:   &body,
		Labels: &labels,
	})
	if err != nil {
		SendErrorMessage(err, h)
		return
	} else {
		SendMessage(fmt.Sprintf("Created issue at %s", *issue.URL), h)
	}

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

	arguments := strings.Split(h.arguments, "$")
	if len(arguments) < 2 {
		msg.Text = "Usage: reminder text $ cron_expression"
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
	val, err := GetChannel(h)
	if err != nil {
		logger.Warnf("Failed to GetChannel channel configuration, %s", h.update.Message.Chat.ID)
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
	val, err := GetChannel(h)
	if err != nil {
		logger.Warnf("Failed to GetChannel channel configuration, %s", h.update.Message.Chat.ID)
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
	val, err := GetChannel(h)
	if err != nil {
		logger.Warnf("Failed to GetChannel channel configuration, %s", h.update.Message.Chat.ID)
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

/* OnLatexMessageHandler handles messages with /tex and then starts a goroutine which would send the document*/
func OnLatexMessageHandler(h MessageHandlerArgs) {
	go LatexTelegramHandler(h)
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
		fmt.Sprintf("Messages so far: <b>%d</b>\nMessages since last request: <b>%d</b>",
			TempChanAttrs[TelegramChannel(h.update.Message.Chat.ID)].MessageCount,
			TempChanAttrs[TelegramChannel(h.update.Message.Chat.ID)].MessageCount-TempChanAttrs[TelegramChannel(h.update.Message.Chat.ID)].LastMessageCountAsked))
	msg.ParseMode = "html"
	TempChanAttrs[TelegramChannel(h.update.Message.Chat.ID)].LastMessageCountAsked =
		TempChanAttrs[TelegramChannel(h.update.Message.Chat.ID)].MessageCount
	logger.Infof(
		"Setting the last asked message count to [%s]",
		TempChanAttrs[TelegramChannel(h.update.Message.Chat.ID)].LastMessageCountAsked,
	)
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

/* OnSedMessageHandler handles messages which starts with /me and converts them to a familiar IRC-like statuses */
func OnSedMessageHandler(h MessageHandlerArgs) {
	if Random.Intn(2) != 1 {
		return
	}
	if len(h.update.Message.Text) > 20 {
		// its a long message. No need to be soo sed for that
		return
	}

	userMsg := strings.Trim(strings.ToLower(h.update.Message.Text), " ")
	msg := tgbotapi.NewMessage(h.update.Message.Chat.ID, "")
	if strings.Contains(userMsg, "indeed") {
		msg.Text = "Very sed indeed."
	} else if userMsg == "sed" {
		msg.Text = "Very sed."
	} else {
		msg.Text = SedResponses[Random.Intn(len(SedResponses))]
	}
	msg.ReplyToMessageID = h.update.Message.MessageID

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

/* OnMessageNotCommandMatchHandler matches those messages which have no associated commands with them */
func OnIdMessageHandler(h MessageHandlerArgs) {
	msg := tgbotapi.NewMessage(
		h.update.Message.Chat.ID, fmt.Sprintf("Id: %d", h.update.Message.Chat.ID))
	msg.ReplyToMessageID = h.update.Message.MessageID
	_, err := h.bot.Send(msg)
	if err != nil {
		logger.Warnf("Couldn't send message without reply to message, %s", err)
	}
}
