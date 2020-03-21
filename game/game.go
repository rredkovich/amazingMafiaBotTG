package game

import (
	"fmt"
	"github.com/rredkovich/amazingMafiaBotTG/types"
)

type Game struct {
	ChatID          int64
	GameInitiator   types.TGUser
	Members         map[string]*types.TGUser
	DeadMembers     map[string]*types.TGUser
	GangsterMembers map[string]*types.TGUser
	Doctor          *types.TGUser
	Commissar       *types.TGUser
	Commands        []InGameCommand
	State           SafeState
	messagesCh      *chan GameMessage
	ticker          Ticker
}

func NewGame(ChatID int64, GameInitiator *types.TGUser, messagesCh *chan GameMessage) *Game {
	// TODO: test for game initiator added as a participant
	members := make(map[string]*types.TGUser)
	members[GameInitiator.UserName] = GameInitiator
	ticker := Ticker{tickStep: 5}

	return &Game{
		ChatID:          ChatID,
		GameInitiator:   *GameInitiator,
		Members:         members,
		DeadMembers:     make(map[string]*types.TGUser),
		GangsterMembers: make(map[string]*types.TGUser),
		messagesCh:      messagesCh,
		ticker:          ticker,
	}
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

		g.SendGroupMessage(text)

		g.ticker.Tick()
	}
}

func (g *Game) Stop() {
	// Do not try to send a message to messagesCh here, will crash the app
	fmt.Printf("Game has been stopped %+v\n", g.ChatID)
	g.State.SetStopped()
}

// Sends message to group, for all players
func (g *Game) SendGroupMessage(msg string) {
	tgMsg := GameMessage{
		ChatID:  g.ChatID,
		Message: msg,
	}
	*g.messagesCh <- tgMsg
}

// Sends message to given user, privatelly
func (g *Game) SendPrivateMessage(msg string, user *types.TGUser) {
	tgMsg := GameMessage{
		ChatID:  int64(user.ID),
		Message: msg,
	}
	*g.messagesCh <- tgMsg
}

// TODO: tests
func (g *Game) UserInGame(u *types.TGUser) bool {
	_, ok := g.Members[u.UserName]
	return ok
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

func (g *Game) KillMember(user *types.TGUser) {
	cmd := InGameCommand{
		Type:   Kill,
		Member: g.Members[user.UserName],
	}

	g.Commands = append(g.Commands, cmd)
}

func (g *Game) HealMember(user *types.TGUser) {
	cmd := InGameCommand{
		Type:   Heal,
		Member: g.Members[user.UserName],
	}

	g.Commands = append(g.Commands, cmd)
}

func (g *Game) InspectMember(user *types.TGUser) {
	cmd := InGameCommand{
		Type:   Inspect,
		Member: g.Members[user.UserName],
	}

	g.Commands = append(g.Commands, cmd)
}

func (g *Game) LynchMember(user *types.TGUser) {
	cmd := InGameCommand{
		Type:   Lynch,
		Member: g.Members[user.UserName],
	}

	g.Commands = append(g.Commands, cmd)
}

func (g *Game) ProcessCommands() {
	nowDead := make(map[string]*types.TGUser)
	for _, cmd := range g.Commands {
		switch cmd.Type {
		case Kill:
			msg := "Мафия выбрала жертву"
			nowDead[cmd.Member.UserName] = cmd.Member
			g.SendGroupMessage(msg)
		case Lynch:
			msg := fmt.Sprintf("Пользователь %+v будет повешен", cmd.Member.UserName)
			nowDead[cmd.Member.UserName] = cmd.Member
			g.SendGroupMessage(msg)
		case Heal:
			msg := "Доктор вышел на ночное дежурство"
			delete(nowDead, cmd.Member.UserName)
			g.SendGroupMessage(msg)
		case Inspect:
			msg := fmt.Sprintf("%+v - %+v", cmd.Member.UserName, cmd.Member.Role)
			g.SendPrivateMessage(msg, g.Commissar)
		}
	}

	for _, user := range nowDead {
		msg := fmt.Sprintf("К сожалению вас убили")
		g.DeadMembers[user.UserName] = user
		g.SendPrivateMessage(msg, user)
	}

	if len(nowDead) == 0 {
		msg := "Удивительно, но этой ночью все выжили"
		g.SendGroupMessage(msg)
	}

}
