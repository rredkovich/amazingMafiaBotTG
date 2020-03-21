package game

import (
	"fmt"
	"github.com/rredkovich/amazingMafiaBotTG/types"
)

type Daytime int

const (
	Day Daytime = iota
	Night
)

type GameMessage struct {
	ChatID  int64
	Message string
}

type Game struct {
	ChatID        int64
	GameInitiator types.TGUser
	Members       map[string]*types.TGUser
	messagesCh    *chan GameMessage
	State         SafeState
	ticker        Ticker
}

func NewGame(ChatID int64, GameInitiator *types.TGUser, messagesCh *chan GameMessage) *Game {
	// TODO: test for game initiator added as a participant
	members := make(map[string]*types.TGUser)
	members[GameInitiator.UserName] = GameInitiator
	ticker := Ticker{tickStep: 5}

	return &Game{
		ChatID:        ChatID,
		GameInitiator: *GameInitiator,
		Members:       members,
		messagesCh:    messagesCh,
		ticker:        ticker,
	}
}

// TODO: tests
func (g *Game) UserInGame(u *types.TGUser) bool {
	_, ok := g.Members[u.UserName]
	return ok
}

func (g *Game) Play() {
	g.State.SetPrepairing()
	g.ticker.RaiseAlarm(10)
	text := fmt.Sprintf("Starting game for %+v", g.ChatID)
	// TODO test that game goroutine exist when game is not Active anymore
	for g.State.IsActive() {
		if g.ticker.Alarm() {
			g.State.SetStarted()
			text = fmt.Sprintf("Game has started %+v", g.ChatID)
			g.State.SetNight()
		}
		fmt.Printf("Playing game for chat %+v\n", g.ChatID)
		msg := GameMessage{g.ChatID, text}
		*g.messagesCh <- msg

		g.ticker.Tick()
	}
}

func (g *Game) Stop() {
	// Do not try to send a message to messagesCh here, will crash the app
	fmt.Printf("Game has been stopped %+v\n", g.ChatID)
	g.State.SetStopped()
}

func (g *Game) UserCouldTalk(userID int) bool {
	// TODO test everybody could talk while gathering people for the game
	if g.State.isStarting {
		return true
	}

	// TODO no one could talk if game is in progress and it is night
	if g.State.GetDayNight() == Night {
		return false
	}

	// TODO logic for person not in members of the game

	return true
	// TODO logic for dead, etc.
}

type CommandType string

const (
	Start                  CommandType = "/start"
	LaunchNewGame                      = "/game"
	EndGame                            = "/endGame"
	ExtendRegistrationTime             = "/extend"
	Leave                              = "/leave"
	Stats                              = "/statistics"
)
