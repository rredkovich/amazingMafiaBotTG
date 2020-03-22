package game

import (
	"errors"
	"github.com/rredkovich/amazingMafiaBotTG/types"
)

type Daytime int

type CommandType string

const (
	Start                  CommandType = "start"
	LaunchNewGame                      = "game"
	EndGame                            = "endGame"
	ExtendRegistrationTime             = "extend"
	Leave                              = "leave"
	Stats                              = "statistics"
)

type GameMessage struct {
	ChatID  int64
	Message string
}

type InGameCommandType int

const (
	Kill InGameCommandType = iota
	Heal
	Lynch
	Inspect
)

type InGameCommand struct {
	Type   InGameCommandType
	Member *types.TGUser
}

// for set data structure
type void struct{}

var AddMemeberGameStarted = errors.New("Игра уже началась, невозможно присоединиться")
var ExtendGameStarted = errors.New("Игра уже началась, невозможно продлить регистрацию")
