package main

import (
	"bufio"
	"context"
	"errors"
	"find_qty/checker"
	"find_qty/m"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	bot    *tgbotapi.BotAPI
	client = &http.Client{Timeout: 10 * time.Second}
)

func main() {
	var err error
	bot, err = tgbotapi.NewBotAPI("<TOKEN>")
	if err != nil {
		// Abort if something is wrong
		log.Panic(err)
	}

	// Set this to true to log all interactions with telegram servers
	bot.Debug = false

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Create a new cancellable background context. Calling `cancel()` leads to the cancellation of the context
	ctx, cancel := context.WithCancel(context.Background())

	// `updates` is a golang channel which receives telegram updates
	updates := bot.GetUpdatesChan(u)
	updates.Clear()
	// Pass cancellable context to goroutine
	go receiveUpdates(ctx, updates, client)

	// Tell the user the bot is online
	log.Println("Start listening for updates. Press enter to stop")

	// Wait for a newline symbol, then cancel handling updates
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	cancel()

}

func receiveUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel, client *http.Client) {
	// `for {` means the loop is infinite until we manually stop it
	for {
		select {
		// stop looping if ctx is cancelled
		case <-ctx.Done():
			return
		// receive update from channel and then handle it
		case update := <-updates:
			handleUpdate(update, client)
		}
	}
}

func handleUpdate(update tgbotapi.Update, client *http.Client) {

	if update.Message != nil {
		handleMessage(update.Message, client)
	}

}

func handleMessage(message *tgbotapi.Message, client *http.Client) {

	user := message.From
	text := message.Text

	if user == nil {
		return
	}

	// Print to console
	log.Printf("%s wrote %s", user.FirstName, text)

	var err error

	if len(text) > 0 {

		answer, err := handleCommand(message.Chat.ID, text, client)

		if err != nil {
			msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
			_, err = bot.Send(msg)

		} else {
			msg := tgbotapi.NewMessage(message.Chat.ID, answer)
			_, err = bot.Send(msg)
		}

	} else {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Пустой ввод")
		_, err = bot.Send(msg)
	}

	if err != nil {
		log.Printf("An error occured: %s", err.Error())
	}

}

// When we get a command, we react accordingly
func handleCommand(chatId int64, command string, client *http.Client) (string, error) {

	cmd := strings.TrimSpace(command)

	switch {
	case cmd == "/help":
		return m.HelpMsg, nil

	case cmd == "/start":
		return m.HelpMsg, nil

	case len(cmd) > 0:
		answer, err := checker.StoreAmount(cmd, client)
		if err != nil {
			return "", err
		}
		return answer, nil
	}

	return "", errors.New("непредвиденная ошибка")
}
