package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/robfig/cron/v3"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

func getJson(url string, target interface{}) error {
	r, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func ScheduleCronFromConfig(config *Config, telegramBot *tgbotapi.BotAPI, c *cron.Cron) {

	for i := range config.VerseSubscribers {
		userId := config.VerseSubscribers[i]
		c.AddFunc(
			fmt.Sprintf("TZ=Asia/Bahrain %s", "00 5 * * *"),
			func() {
				votd := VOTD{}
				err := getJson("https://beta.ourmanna.com/api/v1/get/?format=json", &votd)
				if err != nil {
					logger.Info("Couldn't fetch quote")
					return
				}
				verse := votd.Verse.Details
				logger.Infof("Triggering Daily verse job! in %s", userId)
				msgCron1 := tgbotapi.NewMessage(int64(userId), fmt.Sprintf("<b>Verse of the Day</b>\n\n%s\n~ <i>%s</i> (%s)", verse.Text, verse.Reference, verse.Version))
				msgCron1.ParseMode = "html"

				_, err__ := telegramBot.Send(msgCron1)
				if err__ != nil {
					logger.Warnf("Couldn't send message to channel")
				}

			},
		)

	}

	/*for chanId, chanInstance := range config.Channels {
		chanIdInt := int64(chanId)

		/*c.AddFunc(
			fmt.Sprintf("TZ=%s %s", chanInstance.TimeZone, "59 23 * * *"),
			func() {
				logger.Infof("Triggering Good morning job! in %s", chanId)
				msgCron1 := tgbotapi.NewMessage(chanIdInt, "Good night everyone! 💫💫💫")
				_, err__ := telegramBot.Send(msgCron1)
				if err__ != nil {
					logger.Warnf("Couldn't send message to channel")
				}

			},
		)

		c.AddFunc(
			fmt.Sprintf("TZ=%s %s", chanInstance.TimeZone, "00 7 * * *"),
			func() {
				logger.Infof("Triggering Good morning job! in %s", chanId)
				msgCron1 := tgbotapi.NewMessage(chanIdInt, "Good Morning everyone! 🌤🌤🌤")
				_, err__ := telegramBot.Send(msgCron1)
				if err__ != nil {
					logger.Warnf("Couldn't send message to channel")
				}

			},
		)

		c.AddFunc(
			fmt.Sprintf("TZ=%s %s", chanInstance.TimeZone, "00 8 * * *"),
			func() {
				qod := QuoteOfTheDay{}
				err := getJson("http://quotes.rest/qod", &qod)
				if err != nil {
					logger.Info("Couldn't fetch quote")
					return
				}
				quote := qod.Contents.Quotes[0]
				logger.Infof("Triggering Daily quote job! in %s", chanId)
				msgCron1 := tgbotapi.NewMessage(chanIdInt, fmt.Sprintf("<b>%s</b>\n\n%s\n~ <i>%s</i>", quote.Title, quote.Text, quote.Author))
				msgCron1.ParseMode = "html"

				_, err__ := telegramBot.Send(msgCron1)
				if err__ != nil {
					logger.Warnf("Couldn't send message to channel")
				}

			},
        )
		for k, v := range chanInstance.Reminder {

			instanceId, err := c.AddFunc(
				fmt.Sprintf("TZ=%s %s", chanInstance.TimeZone, k),
				func() {
					logger.Infof("Triggering cronjob set at %s in %s", k, chanId)
					msgCron := tgbotapi.NewMessage(chanIdInt, v.What)
					_, err_ := telegramBot.Send(msgCron)
					if err_ != nil {
						logger.Warnf("Couldn't send message without reply to message, %s", err_)
					}
				})
			if err != nil {
				g := &v
				g.instanceId = int(instanceId)
			}
			logger.Infof("Setting cron at %s, %s with instanceId[%s][%s]", k, err, instanceId, v.What)
        }
	}*/
}
