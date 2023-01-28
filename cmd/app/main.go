package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Wallet map[string]float64

var User = map[int64]Wallet{}

type Binance struct {
	Price float64 `json:price,string`
	Code  int64   `json:code`
}

func main() {
	bot, err := tgbotapi.NewBotAPI("secret")
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			NewMessage(bot, update.Message)
		}
	}
}

func NewMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatId := message.Chat.ID
	s := strings.Split(message.Text, " ")
	switch {
	case s[0] == "ADD":
		if len(s) != 3 {
			message := tgbotapi.NewMessage(chatId, "Ошибка в операции")
			bot.Send(message)
		}
		_, err := getUSD(s[1])
		if err != nil {
			msg := tgbotapi.NewMessage(chatId, "Неверный символ.")
			bot.Send(msg)
		}
		money, err2 := strconv.ParseFloat(s[2], 64)
		if err2 != nil {
			log.Fatal(err2)
		}
		if _, ok := User[chatId]; !ok {
			User[chatId] = make(Wallet)
		}
		User[chatId][s[1]] += money
		message := tgbotapi.NewMessage(chatId, "Валюта добавлена")
		bot.Send(message)

	case s[0] == "SUB":
		if len(s) != 3 {
			message := tgbotapi.NewMessage(chatId, "Ошибка в операции")
			bot.Send(message)
		}
		_, err := getUSD(s[1])
		if err != nil {
			msg := tgbotapi.NewMessage(chatId, "Неверный символ.")
			bot.Send(msg)
		}
		money, err2 := strconv.ParseFloat(s[2], 64)
		if err2 != nil {
			log.Fatal(err2)
		}
		if _, ok := User[chatId]; !ok {
			User[chatId] = make(Wallet)
		}
		User[chatId][s[1]] -= money

		message := tgbotapi.NewMessage(chatId, "Валюта вычтена")
		bot.Send(message)
	case s[0] == "DELETE":
		if len(s) != 2 {
			message := tgbotapi.NewMessage(chatId, "Ошибка в операции")
			bot.Send(message)
		}
		delete(User[chatId], s[1])
		message := tgbotapi.NewMessage(chatId, "Валюта удалена")
		bot.Send(message)
	case s[0] == "SHOW":
		msg := "Баланс $:\n"
		for k, v := range User[chatId] {
			usdprice, err := getUSD(k)
			if err != nil {
				msg := tgbotapi.NewMessage(chatId, err.Error())
				bot.Send(msg)
			}
			msg += fmt.Sprintf("%s: %f\n", k, v*usdprice)
		}
		message := tgbotapi.NewMessage(chatId, msg)
		bot.Send(message)
	default:
		message := tgbotapi.NewMessage(chatId, "Не существует такой команды.")
		bot.Send(message)
	}
}

func getUSD(symbol string) (price float64, err error) {
	URL := fmt.Sprintf("https://www.binance.com/api/v3/ticker/price?symbol=%sUSDT", symbol)
	resp, err := http.Get(URL)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var vBinance Binance
	err = json.NewDecoder(resp.Body).Decode(&vBinance)
	if err != nil {
		return
	}
	if vBinance.Code != 0 {
		err = errors.New("Неверная валюта!")
		return
	}
	return vBinance.Price, nil
}
