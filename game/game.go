package game

import (
	"fmt"
	"github.com/rredkovich/amazingMafiaBotTG/types"
	"math/rand"
	"time"
)

type Game struct {
	ChatID          int64
	ChatTitle       string
	GameInitiator   types.TGUser
	Members         map[string]*types.TGUser
	DeadMembers     map[string]*types.TGUser
	GangsterMembers map[string]*types.TGUser
	Don             *types.TGUser
	Doctor          *types.TGUser
	Commissar       *types.TGUser
	Commands        []InGameCommand
	State           SafeState
	commissarVote   *VoteCommand
	doctorVote      *VoteCommand
	mafiaVote       *VoteCommand
	lynchVote       *VoteCommand
	messagesCh      *chan GameMessage
	votesCh         chan *VoteCommand
	ticker          Ticker
}

func NewGame(ChatID int64, ChatTitle string, GameInitiator *types.TGUser,
	messagesCh *chan GameMessage, votesCh chan *VoteCommand) *Game {
	// TODO: test for game initiator added as a participant
	members := make(map[string]*types.TGUser)
	members[GameInitiator.UserName] = GameInitiator
	ticker := Ticker{tickStep: 1}

	return &Game{
		ChatID:          ChatID,
		ChatTitle:       ChatTitle,
		GameInitiator:   *GameInitiator,
		Members:         members,
		DeadMembers:     make(map[string]*types.TGUser),
		GangsterMembers: make(map[string]*types.TGUser),
		messagesCh:      messagesCh,
		votesCh:         votesCh,
		ticker:          ticker,
	}
}

func (g *Game) Play() {
	g.State.SetPrepairing()
	g.ticker.RaiseAlarm(10)
	//text := fmt.Sprintf("Начинаем мафейку для %+v", g.ChatTitle)
	var text string

	// TODO test that game goroutine exist when game is not Active anymore

	for !g.State.IsStopped() {
		currentState := g.State.GetState()
		if g.ticker.Alarm() {
			if currentState == Preparing {
				// TODO !!! Game should be removed from list of games immideatly after that, memory leak !!!
				if len(g.Members) < 3 {
					g.State.SetStopped()
					g.SendGroupMessage("Слишком мало людей, мафейка не состоится")
					return
				}
				g.AssignRoles()
				g.SendWelcomeMessages()
				g.SendGroupMessage("Игра началась!")
			}
			endingState := g.State.GetState()
			g.finalizeVotingFor(endingState)
			g.ProcessCommands()
			g.State.MoveToNextState()
			currentState = g.State.GetState()
			switch currentState {
			case Day:
				i := rand.Intn(len(dayDescriptions))
				text = fmt.Sprintf("<b>Наступил день</b>\n%+v", dayDescriptions[i])
			case DayVoting:
				// state message handled by day voting
				break
			case Night:
				i := rand.Intn(len(nightDescriptions))
				text = fmt.Sprintf("<b>Наступила ночь</b>\n%+v", nightDescriptions[i])
			}

			if text != "" {
				g.SendGroupMessage(text)
			}
			g.ProcessNewState()
			g.ticker.RaiseAlarm(30)
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

// SendWelcomeMessages sends all users information about their roles
func (g *Game) SendWelcomeMessages() {
	// TODO set Role field in TGUser, use switch-case -> no need for continue here
	for _, user := range g.Members {
		if user == g.Commissar {
			i := rand.Intn(len(commissarDescriptions))
			g.SendPrivateMessage(fmt.Sprintf("Ты коммисар Каттани!\n%+v", commissarDescriptions[i]), user)
			continue
		}

		if user == g.Doctor {
			g.SendPrivateMessage("Ты доктор! Лишь врачебная тайна не дает тебе помочь правосудию. Используй мази и бинты по совести.", user)
			continue
		}

		if user == g.Don {
			g.SendPrivateMessage("Вы - дон, глава мафии, все гангстеры снимают шляпу", user)
			continue
		}

		_, isUserGangsta := g.GangsterMembers[user.UserName]
		if isUserGangsta {
			g.SendPrivateMessage("Вы злодей и душегуб мафиози, доставайте длинный нож", user)
			continue
		}

		i := rand.Intn(len(peaceMemberDescriptions))
		g.SendPrivateMessage(
			fmt.Sprintf("Ты простой мирный житель, %+v\nТвоя задача — вычислить мафию и на городском собрании линчевать засранцев.", peaceMemberDescriptions[i]),
			user)
	}
}

func (g *Game) AddMember(u *types.TGUser) error {
	if g.State.HasStarted() {
		return AddMemeberGameStarted
	}

	_, ok := g.Members[u.UserName]

	if !ok {
		u.Role = "Мирный житель"
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

func (g *Game) UserCouldTalk(userName string) bool {
	// TODO test everybody could talk while gathering people for the game
	if g.State.state == Preparing {
		return true
	}

	// TODO no one could talk if game is in progress and it is night
	if g.State.IsItNight() {
		return false
	}

	// TODO logic for dead
	_, dead := g.DeadMembers[userName]
	if dead {
		return false
	}

	// TODO logic for person not in members of the game
	_, isMember := g.Members[userName]
	if !isMember {
		return false
	}

	return true
}

func (g *Game) AssignRoles() {
	rand.NewSource(time.Now().UnixNano())

	randMax := len(g.Members)
	memberNames := make([]string, 0, len(g.Members))
	for k := range g.Members {
		memberNames = append(memberNames, k)
	}

	doctorInd := rand.Intn(randMax)

	g.Doctor = g.Members[memberNames[doctorInd]]
	g.Doctor.Role = "Доктор"

	// TODO Dirty hack to assign commissar
	g.Commissar = g.Doctor
	for g.Commissar == g.Doctor {
		commissarInd := rand.Intn(randMax)
		g.Commissar = g.Members[memberNames[commissarInd]]
	}
	g.Commissar.Role = "Комиссар Каттани"

	ganstersNum := len(g.Members) / 3

	ganstersAllSet := false
	// TODO Scary shit, could be infinite loop
	for !ganstersAllSet {
		gansterName := memberNames[rand.Intn(randMax)]
		ganster := g.Members[gansterName]

		if ganster == g.Doctor || ganster == g.Commissar {
			continue
		}

		ganster.Role = "Гангстер"
		g.GangsterMembers[ganster.UserName] = ganster

		if len(g.GangsterMembers) == ganstersNum {
			ganstersAllSet = true
		}
	}
}

func (g *Game) KillMember(user *types.TGUser) {
	cmd := InGameCommand{
		Type:   Kill,
		Member: g.Members[user.UserName],
	}

	// kill command should be first, heal last to remove a kill by the heal
	g.Commands = append([]InGameCommand{cmd}, g.Commands...)
}

func (g *Game) HealMember(user *types.TGUser) {
	cmd := InGameCommand{
		Type:   Heal,
		Member: g.Members[user.UserName],
	}

	// heal should be last, kill first to remove a kill by the heal
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
		case Lynch:
			msg := fmt.Sprintf("Пользователь @%+v будет повешен", cmd.Member.UserName)
			nowDead[cmd.Member.UserName] = cmd.Member
			g.SendGroupMessage(msg)
		case Heal:
			delete(nowDead, cmd.Member.UserName)
		case Inspect:
			msg := fmt.Sprintf("@%+v - <b>%+v</b>", cmd.Member.UserName, cmd.Member.Role)
			g.SendPrivateMessage(msg, g.Commissar)
		}
	}

	for _, user := range nowDead {
		g.DeadMembers[user.UserName] = user
		g.SendPrivateMessage("К сожалению вас убили", user)

		g.SendGroupMessage(fmt.Sprintf("@%+v больше нет с нами, прощай <b>%+v</b>", user.UserName, user.Role))
	}

	if len(nowDead) == 0 && len(g.Commands) != 0 {
		msg := "Удивительно, но все выжили"
		g.SendGroupMessage(msg)
	}

	g.Commands = make([]InGameCommand, 0, 2)
}

// ProcessNewState does all logic which should be done on beginning of a State
func (g *Game) ProcessNewState() {
	switch g.State.GetState() {
	case Preparing, Day:
		return
	case DayVoting:
		g.StartVoteLynch()
	case Night:
		g.StartVoteCommissar()
		g.StartVoteDoctor()
		g.StartVoteGansters()
	}
}

func (g *Game) StartVoteGansters() {
	values := make([]*VoteCommandValue, 0, len(g.Members)-len(g.DeadMembers))
	voters := make([]*types.TGUser, 0, len(g.GangsterMembers))
	for _, m := range g.Members {
		// will not add dead members
		_, dead := g.DeadMembers[m.UserName]
		if dead {
			continue
		}
		_, gangster := g.GangsterMembers[m.UserName]
		if gangster {
			values = append(values,
				&VoteCommandValue{fmt.Sprintf("🕴️ %+v", m.UserName), m.UserName})
			voters = append(voters, m)
		} else {
			values = append(values,
				&VoteCommandValue{m.UserName, m.UserName})
		}
	}

	txt := "Пришло твое время, за ниточки будешь дергать ты.\nТебе решать кто не проснётся этой ночью..."
	vcmd := VoteCommand{
		GameChatID:       g.ChatID,
		Action:           StartVoteAction,
		VoteAvailability: VoteAvailabilityEnum.Mafia,
		VoteText:         txt,
		Voters:           voters,
		Values:           values,
	}

	g.mafiaVote = &vcmd
	g.votesCh <- &vcmd
}

var lynchDescriptions = [3]string{
	"У каждого имеется своя судьба, надо только распознать ее. И момент выбора возникает у каждого. Кто же выбрал неверный путь?",
	"Запутавшиеся горожане подозревают друг друга. Но повесить можно только одного. Кого?",
	"В голове головоломка! С кем прекращаем общение?",
}

func (g *Game) StartVoteLynch() {
	values := make([]*VoteCommandValue, 0, len(g.Members)-len(g.DeadMembers))
	voters := make([]*types.TGUser, 0, len(g.Members)-len(g.DeadMembers))
	for _, m := range g.Members {
		// will not add dead members
		_, ok := g.DeadMembers[m.UserName]
		if ok {
			continue
		}
		values = append(values,
			&VoteCommandValue{m.UserName, m.UserName})
		voters = append(voters, m)
	}

	i := rand.Intn(len(lynchDescriptions))
	vcmd := VoteCommand{
		GameChatID:       g.ChatID,
		Action:           StartVoteAction,
		VoteAvailability: VoteAvailabilityEnum.Lynch,
		VoteText:         lynchDescriptions[i],
		Voters:           voters,
		Values:           values,
	}

	g.lynchVote = &vcmd
	g.votesCh <- &vcmd
}

func (g *Game) StartVoteDoctor() {
	if g.DoctorIsDead() {
		return
	}
	//g.SendPrivateMessage("Кого будем лечить?", g.Doctor)

	values := make([]*VoteCommandValue, 0, len(g.Members)-len(g.DeadMembers))
	voters := []*types.TGUser{g.Doctor}

	for _, m := range g.Members {
		// will not add dead members
		_, ok := g.DeadMembers[m.UserName]
		if ok {
			continue
		}
		values = append(values, &VoteCommandValue{m.UserName, m.UserName})
	}

	vcmd := VoteCommand{
		GameChatID:       g.ChatID,
		Action:           StartVoteAction,
		VoteAvailability: VoteAvailabilityEnum.Doctor,
		VoteText:         "Кого забинтуем этой ночью?",
		Voters:           voters,
		Values:           values,
	}

	g.doctorVote = &vcmd
	g.votesCh <- &vcmd
}

func (g *Game) StartVoteCommissar() {
	if g.CommissarIsDead() {
		return
	}

	values := make([]*VoteCommandValue, 0, len(g.Members)-len(g.DeadMembers))
	voters := []*types.TGUser{g.Commissar}

	for _, m := range g.Members {
		// will not add dead members
		_, ok := g.DeadMembers[m.UserName]
		if ok {
			continue
		}
		values = append(values, &VoteCommandValue{m.UserName, m.UserName})
	}

	vcmd := VoteCommand{
		GameChatID:       g.ChatID,
		Action:           StartVoteAction,
		VoteAvailability: VoteAvailabilityEnum.Commissar,
		VoteText:         "Кого проверишь?",
		Voters:           voters,
		Values:           values,
	}

	g.commissarVote = &vcmd
	g.votesCh <- &vcmd
}

func (g *Game) DoctorIsDead() bool {
	_, dead := g.DeadMembers[g.Doctor.UserName]
	return dead
}

func (g *Game) CommissarIsDead() bool {
	_, dead := g.DeadMembers[g.Commissar.UserName]
	return dead
}

func (g *Game) finalizeVotingFor(st State) {
	switch st {
	case Preparing, Day:
		return
	case DayVoting:
		// general voting
		g.lynchVote.Action = StopVoteAction
		g.votesCh <- g.lynchVote
		for g.lynchVote.GetResult() == nil {
			time.Sleep(10 * time.Millisecond)
		}
		lynchResult := g.lynchVote.GetResult()
		g.KillMember(g.Members[lynchResult.Value])
	case Night:
		// doc voting
		g.doctorVote.Action = StopVoteAction
		g.votesCh <- g.doctorVote
		for g.doctorVote.GetResult() == nil {
			time.Sleep(10 * time.Millisecond)
		}
		docResult := g.doctorVote.GetResult()
		g.HealMember(g.Members[docResult.Value])

		// kom voting
		g.commissarVote.Action = StopVoteAction
		g.votesCh <- g.commissarVote
		for g.commissarVote.GetResult() == nil {
			time.Sleep(10 * time.Millisecond)
		}
		comResult := g.commissarVote.GetResult()
		g.InspectMember(g.Members[comResult.Value])

		// mafia voting
		g.mafiaVote.Action = StopVoteAction
		g.votesCh <- g.mafiaVote
		for g.mafiaVote.GetResult() == nil {
			time.Sleep(10 * time.Millisecond)
		}
		mafiaResult := g.mafiaVote.GetResult()
		g.KillMember(g.Members[mafiaResult.Value])
		return
	}
}

func (g *Game) TryToEnd() {
	// cases when game should end:
	// 1. No more gangsters
	// 2. After night only two members are alive
}
