package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/robfig/cron"
)

func ScheduleCronFromConfig(config *Config, telegramBot *tgbotapi.BotAPI, c *cron.Cron){
	for chanId, chanInstance := range config.Channels {
		chanIdInt := int64(chanId)
		for k, v := range chanInstance.Reminder {

			err := c.AddFunc(k, func() {
				logger.Infof("Triggering cronjob set at %s in %s", k, chanId)
				msgCron := tgbotapi.NewMessage(chanIdInt, v.What)
				_, err_ := telegramBot.Send(msgCron)
				if err_ != nil {
					logger.Warnf("Couldn't send message without reply to message, %s", err_)
				}
			})
			logger.Infof("Setting cron at %s, %s", k, err)
		}
	}
}