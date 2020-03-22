package game

import (
	"fmt"
	"github.com/rredkovich/amazingMafiaBotTG/types"
)

type Game struct {
	ChatID          int64
	ChatTitle       string
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

func NewGame(ChatID int64, ChatTitle string, GameInitiator *types.TGUser, messagesCh *chan GameMessage) *Game {
	// TODO: test for game initiator added as a participant
	members := make(map[string]*types.TGUser)
	members[GameInitiator.UserName] = GameInitiator
	ticker := Ticker{tickStep: 5}

	return &Game{
		ChatID:          ChatID,
		ChatTitle:       ChatTitle,
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
	text := fmt.Sprintf("Начинаем мафейку для %+v", g.ChatTitle)

	// TODO test that game goroutine exist when game is not Active anymore

	for !g.State.IsStopped() {
		currentState := g.State.GetState()
		if g.ticker.Alarm() {
			if currentState == Preparing {
				text = "Игра началась!"
				g.SendGroupMessage(text)
			}
			g.State.MoveToNextState()
			currentState = g.State.GetState()
			switch currentState {
			case Day:
				text = "Наступил день"
			case DayVoting:
				text = "Настало время голосовать и наказать засранцев"
			case Night:
				text = "Наступила ночь"
			}

			g.SendGroupMessage(text)
			g.ticker.RaiseAlarm(10)
		}

		g.ticker.Tick()
	}
}

func (g *Game) Stop() {
	// Do not try to send a message to messagesCh here, will crash the app
	fmt.Printf("Game has been stopped %+v\n", g.ChatTitle)
	g.State.SetStopped()
}

func (g *Game) ExtendRegistration(seconds uint) (uint, error) {
	if g.State.HasStarted() {
		return 0, ExtendGameStarted
	}
	g.ticker.PostponeAlarm(seconds)

	return uint(g.ticker.toAlarmValue), nil
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

func (g *Game) AddMember(u *types.TGUser) error {
	if g.State.HasStarted() {
		return AddMemeberGameStarted
	}

	_, ok := g.Members[u.UserName]

	if !ok {
		g.Members[u.UserName] = u
		// Cannot send messages from goroutine while
		/* g.SendPrivateMessage(fmt.Sprintf("Вы вступили в игру в '%+v'", g.ChatTitle), u)
			g.SendGroupMessage(fmt.Sprintf("+1 в игру, теперь нас %+v", len(g.Members)))
		} else {
			 g.SendPrivateMessage("Уже в игрe!", u)
		*/
	}

	return nil
}

// TODO: tests
func (g *Game) UserInGame(u *types.TGUser) bool {
	_, ok := g.Members[u.UserName]
	return ok
}

func (g *Game) UserCouldTalk(userID int) bool {
	// TODO test everybody could talk while gathering people for the game
	if g.State.state == Preparing {
		return true
	}

	// TODO no one could talk if game is in progress and it is night
	if g.State.IsItNight() {
		return false
	}

	// TODO logic for person not in members of the game

	// TODO logic for dead, etc.

	return true
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
			nowDead[cmd.Member.UserName] = cmd.Member
			g.SendGroupMessage("Мафия выбрала жертву")
		case Lynch:
			msg := fmt.Sprintf("Пользователь %+v будет повешен", cmd.Member.UserName)
			nowDead[cmd.Member.UserName] = cmd.Member
			g.SendGroupMessage(msg)
		case Heal:
			delete(nowDead, cmd.Member.UserName)
			g.SendGroupMessage("Доктор вышел на ночное дежурство")
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
