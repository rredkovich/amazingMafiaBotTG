package types

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

type TGUser tgbotapi.User

type Game struct {
	ChatID        int64
	GameInitiator TGUser
}
