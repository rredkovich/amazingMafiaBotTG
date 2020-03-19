package types

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

type RoleType string

const(
	Commissar RoleType = "Комиссар"
	Citizen = "Мирный житель"
	Don = "Дон"
	Gangster = "Гангстер"
	Doctor = "Доктор"
)

type TGUser *tgbotapi.User


type Game struct {
	ChatID int64
	GameInitiator TGUser
}


