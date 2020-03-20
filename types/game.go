package types

import (
	"fmt"
	"sync"
	"time"
)

//type TGUser tgbotapi.User

//type TGUserName string
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

func (s *SafeState) SetPlay() {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.isPlaying = true
	s.isStarting = false
}

func (s *SafeState) SetStart() {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.isPlaying = false
	s.isStarting = true
}

func (s *SafeState) SetStop() {
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

func (s *SafeState) GetDayNight() string {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.dayORNight {
		return "day"
	}

	return "night"
}

type Game struct {
	ChatID        int64
	GameInitiator TGUser
	Members       map[string]*TGUser
	messagesCh    *chan GameMessage
	State         SafeState
}

func NewGame(ChatID int64, GameInitiator *TGUser, messagesCh *chan GameMessage) *Game {
	// TODO: test for game initiator added as a participant
	members := make(map[string]*TGUser)
	members[GameInitiator.UserName] = GameInitiator

	return &Game{
		ChatID:        ChatID,
		GameInitiator: *GameInitiator,
		Members:       members,
		messagesCh:    messagesCh,
	}
}

// TODO: tests
func (g *Game) UserInGame(u *TGUser) bool {
	_, ok := g.Members[u.UserName]
	return ok
}

func (g *Game) Play() {
	g.State.SetStart()
	for g.State.IsActive() {
		fmt.Printf("Playing game for chat %+v\n", g.ChatID)
		msg := GameMessage{g.ChatID, fmt.Sprintf("Playing game for %+v", g.ChatID)}
		*g.messagesCh <- msg
		time.Sleep(5 * time.Second)
	}
}

func (g *Game) Stop() {
	// Do not try to send a message to messagesCh here, will crash the app
	fmt.Printf("Game has been stopped %+v\n", g.ChatID)
	g.State.SetStop()
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
