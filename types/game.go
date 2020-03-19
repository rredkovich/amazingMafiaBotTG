package types

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

type TGUser tgbotapi.User

//type TGUserName string

type Game struct {
	ChatID        int64
	GameInitiator TGUser
	Participants  map[string]*TGUser
}

func NewGame(ChatID int64, GameInitiator *TGUser) *Game {
	// TODO: test for game initiator added as a participant
	participants := make(map[string]*TGUser)
	participants[GameInitiator.UserName] = GameInitiator

	return &Game{
		ChatID:        ChatID,
		GameInitiator: *GameInitiator,
		Participants:  participants,
	}
}

// TODO: tests
func (g *Game) UserInGame(u *TGUser) bool {
	_, ok := g.Participants[u.UserName]
	return ok
}

type CommandType string

const (
	LaunchNewGame          CommandType = "/game"
	Start                              = "/start"
	Stop                               = "/stop"
	ExtendRegistrationTime             = "/extend"
	Leave                              = "/leave"
	Stats                              = "/statistics"
)
