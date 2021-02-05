
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"net/http"
)

type Payload struct {
	Code string `json:"code"`
	Format string `json:"format"`
	Quality int `json:"quality"`
	Density int `json:"density"`
}

type PostResponse struct {
	Status string `json:"status"`
	Filename string `json:"filename"`
	Log string `json:"log,omitempty"`
}

var LatexServerURL = "http://rtex.probablyaweb.site/api/v2"
var TexCleaned = ""


func SendErrorMessage(err error, h MessageHandlerArgs) {
	msg := tgbotapi.NewMessage(h.update.Message.Chat.ID, err.Error()[:4000])
	_, errSend := h.bot.Send(msg)
	logger.Warnf("Err: %s", err)
	if errSend != nil {
		logger.Warnf("Couldn't send message to Telegram chat, %s", errSend)
	}
}


func LatexTelegramHandler(h MessageHandlerArgs) {
	texUrl, err := GetLatexImage(h.arguments)
	if err != nil {
		SendErrorMessage(err, h)
		return
	}

	res, err := http.Get(texUrl)

	if err != nil {
		SendErrorMessage(err, h)
		return
	}

	content, err := ioutil.ReadAll(res.Body)

	if err != nil {
		// error handling...
	}

	texBytes := tgbotapi.FileBytes{Name: "image.jpg", Bytes: content}
	msg := tgbotapi.NewPhotoUpload(h.update.Message.Chat.ID, texBytes)


	_, errSend := h.bot.Send(msg)
	if errSend != nil {
		logger.Warnf("Couldn't send message to Telegram chat, %s", errSend)
	}
}



func GetLatexImage(arguments string) (string, error) {
	var payloadJson []byte

	payload := Payload{
		Code: fmt.Sprintf(TexTemplate, arguments),
		Format: "png",
		Quality: 80,
		Density: 300,
	}

	logger.Info("Sending payload:", payload)
	payloadJson, err := json.Marshal(&payload)
	if err != nil {
		return "", err
	}

	logger.Info("Latex - Sending post request")
	r, err := http.Post(LatexServerURL, "application/json", bytes.NewBuffer(payloadJson))
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	logger.Info("Received post request, processing json")
	rawPostResponse, err := ioutil.ReadAll(r.Body)
	var postResponse PostResponse
	err = json.Unmarshal(rawPostResponse, &postResponse)
	if err != nil  {
		return "", err
	}
	if postResponse.Status != "success" {
		return "", errors.New(fmt.Sprintf("latex couldn't be processed. %s, %s", postResponse, rawPostResponse))
	}

	logger.Info("Sending get request to get the filename")
	getRequestUrl := fmt.Sprintf("%s/%s", LatexServerURL, postResponse.Filename)

	return getRequestUrl, nil

}