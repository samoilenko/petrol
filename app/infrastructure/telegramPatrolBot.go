package infrastructure

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type TelegramToken string

type TelegramBot struct {
	Bot     *tgbotapi.BotAPI
	chatIds []int64
}

func (t *TelegramBot) Start() {
	updates := t.Bot.ListenForWebhook("/telegram/bot")

	fmt.Println("Telegram bot has started")

	for update := range updates {
		fmt.Printf("CHANT ID %d", update.Message.Chat.ID)
		_, err := t.Bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Здоровеньки були"))

		if err != nil {
			fmt.Println(err)
		}

		t.chatIds = append(t.chatIds, update.Message.Chat.ID)
	}
}

func (t *TelegramBot) Inform(message string) []error {
	errs := make([]error, 0)

	for _, chatID := range t.chatIds {
		_, err := t.Bot.Send(tgbotapi.NewMessage(
			chatID,
			message,
		))
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func NewTelegramBot(token TelegramToken, webhookURL string, initChatIds []int64, debug bool) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(string(token))
	bot.Debug = debug
	if err != nil {
		return nil, err
	}

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(webhookURL))

	if err != nil {
		return nil, err
	}

	fmt.Println(fmt.Sprintf("Telegram bot is authorized on account %s", bot.Self.UserName))

	return &TelegramBot{
		Bot:     bot,
		chatIds: initChatIds,
	}, nil
}
