package types

import (
	"fmt"
	"sync"
	"time"
)

type Daytime int

const (
	Day Daytime = iota
	Night
)

type TGUser struct {
	ID        int
	UserName  string
	FirstName string
	LastName  string
}

func NewTGUser(id int, userName string, firstName string, lastName string) *TGUser {
	return &TGUser{
		ID:        id,
		UserName:  userName,
		FirstName: firstName,
		LastName:  lastName,
	}
}

type GameMessage struct {
	ChatID  int64
	Message string
}

type SafeState struct {
	isPlaying  bool
	isStarting bool
	dayORNight bool //true - day, false - night
	mux        sync.Mutex
}

func (s *SafeState) SetStarted() {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.isPlaying = true
	s.isStarting = false
}

func (s *SafeState) SetPrepairing() {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.isPlaying = false
	s.isStarting = true
}

func (s *SafeState) SetStopped() {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.isPlaying = false
	s.isStarting = false
}

func (s *SafeState) IsActive() bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	return s.isPlaying || s.isStarting

}

func (s *SafeState) SetDay() {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.dayORNight = true
}

func (s *SafeState) SetNight() {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.dayORNight = false
}

func (s *SafeState) GetDayNight() Daytime {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.dayORNight {
		return Day
	}

	return Night
}

type Game struct {
	ChatID        int64
	GameInitiator TGUser
	Members       map[string]*TGUser
	messagesCh    *chan GameMessage
	State         SafeState
	ticker        Ticker
}

type Ticker struct {
	mux                  sync.Mutex
	value                uint32
	tickStep             uint32
	toAlarmValue         int32
	lastBeforeAlarmValue int32
}

func (t *Ticker) Alarm() bool {
	t.mux.Lock()
	defer t.mux.Unlock()

	return t.lastBeforeAlarmValue <= 0
}

func (t *Ticker) Tick() {
	t.mux.Lock()
	defer t.mux.Unlock()

	t.value += t.tickStep
	if t.toAlarmValue > 0 {
		t.lastBeforeAlarmValue -= int32(t.tickStep)
	}
	time.Sleep(time.Duration(t.tickStep) * time.Second)
}

func (t *Ticker) RaiseAlarm(seconds uint32) {
	t.mux.Lock()
	defer t.mux.Unlock()

	t.toAlarmValue = int32(seconds) // could be negative if tickStep != 1
	t.lastBeforeAlarmValue = int32(seconds)
}

func (t *Ticker) GetValue() uint32 {
	t.mux.Lock()
	defer t.mux.Unlock()

	return t.value
}

func NewGame(ChatID int64, GameInitiator *TGUser, messagesCh *chan GameMessage) *Game {
	// TODO: test for game initiator added as a participant
	members := make(map[string]*TGUser)
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
func (g *Game) UserInGame(u *TGUser) bool {
	_, ok := g.Members[u.UserName]
	return ok
}

func (g *Game) Play() {
	g.State.SetPrepairing()
	g.ticker.RaiseAlarm(10)
	text := fmt.Sprintf("Starting game for %+v", g.ChatID)
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
