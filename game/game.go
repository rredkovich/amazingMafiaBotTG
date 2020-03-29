package game

import (
	"fmt"
	"github.com/rredkovich/amazingMafiaBotTG/types"
	"log"
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
	r               *rand.Rand
}

func NewGame(ChatID int64, ChatTitle string, GameInitiator *types.TGUser,
	messagesCh *chan GameMessage, votesCh chan *VoteCommand) *Game {
	members := make(map[string]*types.TGUser)
	ticker := Ticker{tickStep: 1}

	g := Game{
		ChatID:          ChatID,
		ChatTitle:       ChatTitle,
		GameInitiator:   *GameInitiator,
		Members:         members,
		DeadMembers:     make(map[string]*types.TGUser),
		GangsterMembers: make(map[string]*types.TGUser),
		messagesCh:      messagesCh,
		votesCh:         votesCh,
		ticker:          ticker,
		r:               rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	g.assignBaseRole(&g.GameInitiator)
	g.Members[GameInitiator.UserName] = &g.GameInitiator

	return &g
}

func (g *Game) Play(prepareTime uint32) {

	g.State.SetPrepairing()
	g.ticker.RaiseAlarm(prepareTime)

	// TODO test that game goroutine exist when game is not Active anymore

	for !g.State.IsStopped() {
		currentState := g.State.GetState()
		if g.ticker.AlarmIsSoon() {
			switch currentState {
			case Preparing:
				g.SendGroupMessage("Игра начинается через 30 секунд")
			case DayVoting:
				g.SendGroupMessage("Голосование закончится через 30 секунд")
			}
		}

		if g.ticker.Alarm() {
			if currentState == Preparing {
				// TODO !!! Game should be removed from list of games in main thread immideatly after that, memory leak !!!
				if len(g.Members) < 4 {
					g.State.SetStopped()
					g.SendGroupMessage("Слишком мало людей, мафейка не состоится")
					return
				}
				g.AssignRoles()
				g.SendWelcomeMessages()
				g.SendGroupMessage("<b>Игра началась!</b>")
			}
			endingState := g.State.GetState()
			log.Printf("Ending state %+v for game %+v ", g.State.GetState(), g.ChatID)
			g.finalizeVotingFor(endingState)
			g.ProcessCommands()
			ended, duration := g.TryToEnd()

			if ended {
				results := g.GetResults()
				g.SendGroupMessage(fmt.Sprintf("%+v\nИгра длилась %+v секунд", results, duration))
				return
			}

			g.State.MoveToNextState()
			currentState = g.State.GetState()
			switch currentState {
			case Day:
				i := g.r.Intn(len(dayDescriptions))
				g.SendGroupMessage(fmt.Sprintf("<b>Наступил день</b>\n%+v", dayDescriptions[i]))
				g.SendGroupMessage(g.ListAlive())
			case DayVoting:
				// state message handled by day voting
				break
			case Night:
				i := g.r.Intn(len(nightDescriptions))
				g.SendGroupMessage(fmt.Sprintf("<b>Наступила ночь</b>\n%+v", nightDescriptions[i]))
			}

			g.ProcessNewState()
			g.ticker.RaiseAlarm(60)
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
	for _, user := range g.Members {
		if user == g.Commissar {
			i := g.r.Intn(len(commissarDescriptions))
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

		i := g.r.Intn(len(peaceMemberDescriptions))
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
		g.assignBaseRole(u)
		g.Members[u.UserName] = u
		// Cannot send messages from goroutine while

		// shitty duck taped solution
		//go func() {g.SendPrivateMessage(fmt.Sprintf("Вы вступили в игру в '%+v'", g.ChatTitle), u)}()

		/* g.SendPrivateMessage(fmt.Sprintf("Вы вступили в игру в '%+v'", g.ChatTitle), u)
			g.SendGroupMessage(fmt.Sprintf("+1 в игру, теперь нас %+v", len(g.Members)))
		} else {
			 g.SendPrivateMessage("Уже в игрe!", u)
		*/
	}

	return nil
}

func (g *Game) UserInGame(u *types.TGUser) bool {
	_, ok := g.Members[u.UserName]
	return ok
}

// TODO: test UserCouldTalk
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

// TODO test ProcessInGameDirectMessage
func (g *Game) ProcessInGameDirectMessage(userName string, msg string) {
	_, inGame := g.Members[userName]
	if !inGame {
		return
	}

	u, dead := g.DeadMembers[userName]
	if dead {
		g.DeadWantsToTalk(u, msg)
	}
}

// TODO test DeadWantsToTalk
func (g *Game) DeadWantsToTalk(u *types.TGUser, msg string) {
	if !u.SpokenLastWords {
		i := g.r.Intn(len(deadNotes))
		notePrefix := fmt.Sprintf(deadNotes[i], u.InlineTGName())
		note := fmt.Sprintf("%+v <i>%+v</i>", notePrefix, msg)
		// TODO bad could leak memory
		u.SpokenLastWords = true
		go func() { g.SendGroupMessage(note) }()
		return
	}

	/*
		// TODO bad could leak memory, test thoughfuly before release
		// Sends dead messages from one dead to another
		go func() {
			time.Sleep(10 * time.Millisecond)
			for _, du := range g.DeadMembers {
				if du == u {
					return
				}

			words := fmt.Sprintf("%+v: %+v", u.InlineTGName(), msg)
			g.SendPrivateMessage(words, u)
		}()
	*/

}

func (g *Game) AssignRoles() {
	randMax := len(g.Members)
	memberNames := make([]string, 0, len(g.Members))
	for k, member := range g.Members {
		member.Role = types.Citizen
		memberNames = append(memberNames, k)
	}

	doctorInd := g.r.Intn(randMax)

	g.Doctor = g.Members[memberNames[doctorInd]]
	g.Doctor.Role = types.Doctor

	// no commissar for games less of 4 members
	if len(g.Members) > 4 {
		// TODO Dirty hack to assign commissar
		g.Commissar = g.Doctor
		for g.Commissar == g.Doctor {
			commissarInd := g.r.Intn(randMax)
			g.Commissar = g.Members[memberNames[commissarInd]]
		}
		g.Commissar.Role = types.Commissar
	}

	ganstersNum := len(g.Members) / 3

	ganstersAllSet := false
	// TODO Scary shit, could be infinite loop while setting gangsters
	for !ganstersAllSet {
		gansterName := memberNames[g.r.Intn(randMax)]
		ganster := g.Members[gansterName]

		if ganster == g.Doctor || ganster == g.Commissar {
			continue
		}

		ganster.Role = types.Gangster
		g.GangsterMembers[ganster.UserName] = ganster

		if len(g.GangsterMembers) == ganstersNum {
			ganstersAllSet = true
		}
	}
}

// ExecuteKill marks a member to be killed if command is not nil
func (g *Game) ExecuteKill(value *VoteCommandValue) {
	if value == nil {
		return
	}

	cmd := InGameCommand{
		Type:   Kill,
		Member: g.Members[value.Value],
	}

	// kill command should be first, heal last to remove a kill by the heal
	g.Commands = append([]InGameCommand{cmd}, g.Commands...)
}

// ExecuteHeal marks a member to be healed if command is not nil
func (g *Game) ExecuteHeal(value *VoteCommandValue) {
	if value == nil {
		return
	}

	cmd := InGameCommand{
		Type:   Heal,
		Member: g.Members[value.Value],
	}

	// heal should be last, kill first to remove a kill by the heal
	g.Commands = append(g.Commands, cmd)
}

// ExecuteInspect marks member to be inspected if command is not nil
func (g *Game) ExecuteInspect(value *VoteCommandValue) {
	if value == nil {
		return
	}

	cmd := InGameCommand{
		Type:   Inspect,
		Member: g.Members[value.Value],
	}

	g.Commands = append(g.Commands, cmd)
}

// ExecuteLynch marks member to be lynched if command is not nil
func (g *Game) ExecuteLynch(value *VoteCommandValue) {
	if value == nil {
		return
	}

	cmd := InGameCommand{
		Type:   Lynch,
		Member: g.Members[value.Value],
	}

	g.Commands = append(g.Commands, cmd)
}

func (g *Game) ProcessCommands() {
	lynchVote := false
	nowDead := make(map[string]*types.TGUser)
	for _, cmd := range g.Commands {
		switch cmd.Type {
		case Kill:
			nowDead[cmd.Member.UserName] = cmd.Member
		case Lynch:
			nowDead[cmd.Member.UserName] = cmd.Member
			lynchVote = true
		case Heal:
			delete(nowDead, cmd.Member.UserName)
			g.SendPrivateMessage("Доктор приходил к вам", cmd.Member)
		case Inspect:
			msg := fmt.Sprintf("%+v - <b>%+v</b>", cmd.Member.InlineTGName(), cmd.Member.Role)
			g.SendPrivateMessage("Кто-то заинтересовался вашей ролью", cmd.Member)
			g.SendPrivateMessage(msg, g.Commissar)
		}
	}

	for _, user := range nowDead {
		g.DeadMembers[user.UserName] = user

		if g.Commissar == user {
			// TODO should special roles players be nilified after death?..
			g.Commissar = nil
			g.commissarVote = nil
		}

		if g.Doctor == user {
			g.Doctor = nil
			g.doctorVote = nil
		}

		if g.Don == user {
			g.Don = nil
		}

		if lynchVote {
			g.SendPrivateMessage("Тебя линчевали на дневном собрании 😞", user)
			i := g.r.Intn(len(lynchResults))
			msg := fmt.Sprintf(lynchResults[i], user.InlineTGName()) + fmt.Sprintf("\n\nРоль - <b>%+v</b>", user.Role)
			g.SendGroupMessage(msg)
		} else {
			i := g.r.Intn(len(deaths))
			askForDeadNote := fmt.Sprintf(deadNotePrompt, deaths[i])
			g.SendPrivateMessage(askForDeadNote, user)
			// TODO comissar kill should be with different text
			j := g.r.Intn(len(mafiaDeathsDescriptions))
			g.SendGroupMessage(fmt.Sprintf(mafiaDeathsDescriptions[j], user.InlineTGName(), user.Role))
		}

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
				&VoteCommandValue{fmt.Sprintf("🕴️ %+v", m.HumanReadableName()), m.UserName})
			voters = append(voters, m)
		} else {
			values = append(values,
				&VoteCommandValue{m.HumanReadableName(), m.UserName})
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
			&VoteCommandValue{m.HumanReadableName(), m.UserName})
		voters = append(voters, m)
	}

	i := g.r.Intn(len(lynchDescriptions))
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

// TODO tests game doesn't crash on vote when g.Doctor == nil
func (g *Game) StartVoteDoctor() {
	if g.Doctor == nil || g.DoctorIsDead() {
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
		values = append(values, &VoteCommandValue{m.HumanReadableName(), m.UserName})
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

// TODO tests game doesn't crash on vote when g.Commissar == nil
func (g *Game) StartVoteCommissar() {
	if g.Commissar == nil || g.CommissarIsDead() {
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
		values = append(values, &VoteCommandValue{m.HumanReadableName(), m.UserName})
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
	if g.Doctor == nil {
		return true
	}
	_, dead := g.DeadMembers[g.Doctor.UserName]
	return dead
}

func (g *Game) CommissarIsDead() bool {
	if g.Commissar == nil {
		return true
	}
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
		for {
			_, lynchSynced := g.lynchVote.GetResult()

			log.Printf("Votes are ready lynch: %+v", lynchSynced)
			if lynchSynced {
				break
			}
			time.Sleep(30 * time.Millisecond)
		}

		lynchResult, _ := g.lynchVote.GetResult()
		g.ExecuteLynch(lynchResult)
	case Night:
		g.EndVoteCommissar()
		g.EndVoteGansters()
		g.EndVoteDoctor()
	}
}

func (g *Game) EndVoteCommissar() {
	if g.Commissar == nil || g.CommissarIsDead() {
		return
	}
	g.commissarVote.Action = StopVoteAction
	g.votesCh <- g.commissarVote

	for {
		_, comSynced := g.commissarVote.GetResult()

		log.Printf("Votes are ready commissar: %+v", comSynced)
		if comSynced {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	comResult, _ := g.commissarVote.GetResult()
	g.ExecuteInspect(comResult)
}

func (g *Game) EndVoteDoctor() {
	if g.Doctor == nil || g.DoctorIsDead() {
		return
	}
	g.doctorVote.Action = StopVoteAction
	g.votesCh <- g.doctorVote

	for {
		_, docSynced := g.doctorVote.GetResult()

		log.Printf("Votes are ready doctor: %+v", docSynced)
		if docSynced {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	docResult, _ := g.doctorVote.GetResult()
	g.ExecuteHeal(docResult)
}

func (g *Game) EndVoteGansters() {
	g.mafiaVote.Action = StopVoteAction
	g.votesCh <- g.mafiaVote

	for {
		_, mafSynced := g.mafiaVote.GetResult()

		log.Printf("Votes are ready mafia: %+v", mafSynced)
		if mafSynced {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// mafia voting
	mafiaResult, _ := g.mafiaVote.GetResult()
	g.ExecuteKill(mafiaResult)
}

func (g *Game) TryToEnd() (bool, uint32) {
	// cases when game should end:
	// 1. Main. After night only two members are alive
	if len(g.Members)-len(g.DeadMembers) <= 2 {
		g.Stop()
		return true, g.ticker.GetValue()
	}

	shouldStop := true
	// 1. No more gangsters
	for gangsterUsername := range g.GangsterMembers {
		_, dead := g.DeadMembers[gangsterUsername]
		if !dead {
			shouldStop = false
			break
		}
	}

	if shouldStop {
		g.Stop()
	}

	return shouldStop, g.ticker.GetValue()
}

func (g *Game) GetResults() string {
	alive := make([]*types.TGUser, 0, len(g.Members)-len(g.DeadMembers))
	aliveGansters := make([]*types.TGUser, 0, 2)

	winnerText := "Все мужчины клана погибли в вендеттах. Город освобожден."
	winnersList := "Победители:\n"
	defeatedList := "Проигравшие:\n"

	for _, member := range g.Members {
		_, dead := g.DeadMembers[member.UserName]
		if !dead {
			_, gangster := g.GangsterMembers[member.UserName]
			if gangster {
				aliveGansters = append(aliveGansters, member)
			} else {
				alive = append(alive, member)
			}
		} else {
			defeatedList += fmt.Sprintf("  - %+v - <b>%+v</b>\n", member.InlineTGName(), member.Role)
		}
	}

	if len(aliveGansters) != 0 {
		for _, gangster := range aliveGansters {
			winnersList += fmt.Sprintf("  - %+v - <b>%+v</b>\n", gangster.InlineTGName(), gangster.Role)
			winnerText = "Город захвачен сицилийскими псинами сутулыми. Мафия победила."
		}

		// mafia won, all alive member have loosed
		for _, member := range alive {
			defeatedList += fmt.Sprintf("  - %+v - <b>%+v</b>\n", member.InlineTGName(), member.Role)
		}
	} else {
		for _, member := range alive {
			winnersList += fmt.Sprintf("  - %+v - <b>%+v</b>\n", member.InlineTGName(), member.Role)
		}
	}

	text := fmt.Sprintf("<b>Игра завершена</b>\n%+v \n%+v \n%+v", winnerText, winnersList, defeatedList)
	return text
}

func (g *Game) MembersCount() int {
	return len(g.Members)
}

func (g *Game) ListAlive() string {
	text := "Живые игроки:\n"
	var roles []types.RoleType
	for _, member := range g.Members {
		_, dead := g.DeadMembers[member.UserName]
		if !dead {
			text += fmt.Sprintf("- %+v\n", member.InlineTGName())
			roles = append(roles, member.Role)
		}
	}

	g.r.Seed(time.Now().UnixNano())
	g.r.Shuffle(len(roles), func(i, j int) { roles[i], roles[j] = roles[j], roles[i] })

	text += "\n<b>Роли:</b> "
	for _, role := range roles {
		text += fmt.Sprintf("%+v, ", role)
	}

	// cutting last ,
	return text[:len(text)-1]
}

// assignBaseRole set base citizen role for member
func (g *Game) assignBaseRole(m *types.TGUser) {
	m.Role = types.Citizen
}
