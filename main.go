package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rredkovich/amazingMafiaBotTG/game"
	"github.com/rredkovich/amazingMafiaBotTG/types"
	"github.com/segmentio/ksuid"
	"log"
	"os"
	"strconv"
)

type TGVoteValue struct {
	UUID   ksuid.KSUID
	VoteID string
	Value  string
	User   *tgbotapi.User
}

func NewTGVoteValue(voteID string, value string) *TGVoteValue {
	uuid := ksuid.New()
	return &TGVoteValue{
		VoteID: voteID,
		Value:  value,
		UUID:   uuid,
	}
}

func (v *TGVoteValue) UUIDString() string {
	return v.UUID.String()
}

type Vote struct {
	ID               string
	GameChatID       int64
	VoteText         string
	Action           game.VoteAction
	VoteAvailability game.VoteAvailability
	Voters           []*types.TGUser
	Values           []*game.VoteCommandValue
	//Votes map[*types.TGUser]*game.VoteCommandValue
	Votes map[*tgbotapi.User]string
}

func NewVoteFromVoteCommand(vcmd *game.VoteCommand, ID string) *Vote {
	return &Vote{
		ID,
		vcmd.GameChatID,
		vcmd.VoteText,
		vcmd.Action,
		vcmd.VoteAvailability,
		vcmd.Voters,
		vcmd.Values,
		make(map[*tgbotapi.User]string),
	}
}

// StartVote return map of ids of chats and messages which should be sent
func (v *Vote) StartVote(votes map[string]*TGVoteValue) []*tgbotapi.MessageConfig {
	switch v.VoteAvailability {
	case game.VoteAvailabilityEnum.Lynch:
		messages := make([]*tgbotapi.MessageConfig, 0, len(v.Voters)+1) // +1 for group chat message

		// message to group chat regarding lynch
		groupMsg := tgbotapi.NewMessage(v.GameChatID, v.VoteText)
		kbd := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("Голосовать", "https://t.me/amafia_bot"),
			),
		)
		groupMsg.ReplyMarkup = kbd
		messages = append(messages, &groupMsg)

		// individual messages with buttons for every voter
		for _, voter := range v.Voters {
			kbdRows := make([][]tgbotapi.InlineKeyboardButton, 0, len(v.Values)-1) // will not vote for himself
			for _, value := range v.Values {
				// will note vote for himself
				if value.Value == voter.UserName {
					continue
				}

				vote := NewTGVoteValue(v.ID, value.Value)
				voteUUID := vote.UUIDString()
				votes[voteUUID] = vote
				key := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(value.Text, voteUUID)}
				kbdRows = append(kbdRows, key)
			}
			msg := tgbotapi.NewMessage(int64(voter.ID), "Кого желаем вздернуть?")
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(kbdRows...)
			messages = append(messages, &msg)
		}

		return messages

	case game.VoteAvailabilityEnum.Mafia:
		messages := make([]*tgbotapi.MessageConfig, 0, len(v.Voters))

		// individual messages with buttons for every voter
		for _, voter := range v.Voters {
			kbdRows := make([][]tgbotapi.InlineKeyboardButton, 0, len(v.Values)-1) // will not vote for himself
			for _, value := range v.Values {
				// will note vote for himself
				// value.Text here has mafia emoji
				// TODO operate with userID's everywhere
				if value.Value == voter.UserName {
					continue
				}
				vote := NewTGVoteValue(v.ID, value.Value)
				voteUUID := vote.UUIDString()
				votes[voteUUID] = vote
				row := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(value.Text, voteUUID)}
				kbdRows = append(kbdRows, row)
			}
			msg := tgbotapi.NewMessage(int64(voter.ID), "Кого желаем вздернуть?")
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(kbdRows...)
			messages = append(messages, &msg)
		}

		return messages
	}
	return nil
}

func (v *Vote) RegisterVote(u *tgbotapi.User, vote string) error {
	// only voters could vote
	fail := true
	for _, voter := range v.Voters {
		if voter.ID == u.ID {
			fail = false
			break
		}
	}
	if fail {
		return errors.New("Голосование не для вас")
	}

	v.Votes[u] = vote
	return nil
}

func (v *Vote) EndVote() (*game.VoteCommandValue, error) {
	counters := make(map[string]int)

	for _, vote := range v.Votes {
		cntr, ok := counters[vote]
		if !ok {
			counters[vote] = 1
		} else {
			counters[vote] = cntr + 1
		}
	}

	finalVote := ""
	greaterCntr := 0

	for vote, cntr := range counters {
		if cntr > greaterCntr {
			greaterCntr = cntr
			finalVote = vote
		}
	}

	for _, value := range v.Values {
		if value.Value == finalVote {
			return value, nil
		}
	}

	return nil, errors.New("Cannot get result of a vote!!!")
}

func main() {

	games := make(map[int64]*game.Game)
	votes := make(map[string]*Vote)
	tgVotes := make(map[string]*TGVoteValue)

	messagesFromGames := make(chan game.GameMessage)
	voteCommandsFromGames := make(chan *game.VoteCommand)

	fmt.Println("It's a mafia bot")
	fmt.Println(types.Doctor)

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TG_API_TOKEN"))
	if err != nil {
		panic(err) // You should add better error handling than this!
	}

	bot.Debug = true // Has the library display every request and response.

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 10

	updatesCh, err := bot.GetUpdatesChan(u)

	for {
		select {
		case cmd := <-voteCommandsFromGames:
			voteKey := fmt.Sprintf("%+v_%v", cmd.GameChatID, cmd.VoteAvailability)
			switch cmd.Action {
			case game.StartVoteAction:
				fmt.Printf("Got vote command %+v\n", cmd)
				vote := NewVoteFromVoteCommand(cmd, voteKey)
				votes[voteKey] = vote
				voteMsgs := vote.StartVote(tgVotes)
				for _, voteMsg := range voteMsgs {
					_, err := bot.Send(voteMsg)
					if err != nil {
						log.Printf("Cannot send vote message! \n%+v\n%+v\n", voteMsg, err)
					}
				}
			case game.StopVoteAction:
				vote := votes[voteKey]
				result, err := vote.EndVote()
				if err != nil {
					fmt.Printf("Vote step into shit %+v", err)
					delete(votes, voteKey)
					continue
				}
				cmd.SetResult(result)
				delete(votes, voteKey)
			}

		case msg := <-messagesFromGames:
			log.Printf("Games: %+v\n", games)
			fmt.Printf("Got GameMessage %+v\n", msg)
			tgMsg := tgbotapi.NewMessage(msg.ChatID, msg.Message)
			tgMsg.ParseMode = "html"

			_, err = bot.Send(tgMsg)

			if err != nil {
				log.Printf("Got error on send! %+v\n", err)
			}
		case update := <-updatesCh:
			log.Printf("Games: %+v\n", games)
			// command execution from voting
			if update.Message == nil {
				log.Printf("VOTE ENUM %s", game.VoteActionEnum.Start)
				log.Printf("%+v\n", update.CallbackQuery)
				tgVote, ok := tgVotes[update.CallbackQuery.Data]
				if !ok {
					fmt.Printf("Cannot get vote by callback query '%+v'", update.CallbackQuery.Data)
					continue
				}
				vote := votes[tgVote.VoteID]

				var answerText string
				err := vote.RegisterVote(update.CallbackQuery.From, tgVote.Value)
				if err != nil {
					answerText = fmt.Sprintf("%+v", err)
				} else {
					answerText = "Ваш голос учтен"

				}
				bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "👌"))
				bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, answerText))
				continue
			} // ignore any non-Message Updates

			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			switch update.Message.IsCommand() {
			case true:
				switch game.CommandType(update.Message.Command()) {
				case game.LaunchNewGame:
					gameID := update.Message.Chat.ID

					// TODO /test cannot start game if started
					// TODO /notify "game already started" if '/game' received again
					g, ok := games[gameID]
					if ok && !g.State.IsStopped() {
						tgMsg := tgbotapi.NewMessage(g.ChatID, "Уже есть активная игра")
						_, err = bot.Send(tgMsg)
						continue
					}

					if ok && g.State.IsStopped() {
						delete(games, gameID)
					}

					gameIDStr := strconv.AppendInt([]byte(""), gameID, 10)
					encodedGameID := base64.StdEncoding.EncodeToString([]byte(gameIDStr))

					kbd := tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonURL(
								"Присоединиться",
								fmt.Sprintf("https://t.me/amafia_bot?start=%+v", encodedGameID),
							),
						),
					)

					from := update.Message.From
					starter := types.NewTGUser(from.ID, from.UserName, from.FirstName, from.LastName)
					game := game.NewGame(update.Message.Chat.ID, update.Message.Chat.Title, starter,
						&messagesFromGames, voteCommandsFromGames)
					games[game.ChatID] = game
					go game.Play()
					log.Printf("Created game: %+v\n", game)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ведется набор в игру")
					msg.ReplyMarkup = kbd
					bot.Send(msg)

				case game.EndGame:
					game, ok := games[update.Message.Chat.ID]
					if !ok {
						continue
					}
					// TODO test game removed from games when '/endGame' received
					delete(games, game.ChatID)
					game.Stop()
					log.Printf("Stopped game: %+v\n", game)

					tgMsg := tgbotapi.NewMessage(game.ChatID, "Игра остановлена")

					_, err = bot.Send(tgMsg)

					if err != nil {
						log.Printf("Got error on send! %+v\n", err)
					}
				case game.Start:
					// Expecting new user trying to add himself to the game
					//"/start -1001256382007"
					if update.Message.Text == "/start" {
						continue
					}
					receivedEncodedGameID := update.Message.Text[7:]
					var decodedByte, _ = base64.StdEncoding.DecodeString(receivedEncodedGameID)
					decodedGameID, err := strconv.ParseInt(string(decodedByte), 10, 64)
					if err != nil {
						fmt.Printf("Cannot parce game ID on start: %+v\n", err)
						continue
					}
					fmt.Printf("Got game id %+v\n", decodedGameID)
					startedGame, ok := games[decodedGameID]
					if !ok {
						gameDead := tgbotapi.NewMessage(update.Message.Chat.ID, "Игры нет, возможно закончилась?..")
						bot.Send(gameDead)
						continue
					}

					from := update.Message.From
					err = startedGame.AddMember(types.NewTGUser(from.ID, from.UserName, from.FirstName, from.LastName))
					var answer tgbotapi.MessageConfig
					if err != nil {
						answer = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("%+v", err))

					} else {
						answer = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Вы присоединились к игре в <b>%+v</b>", startedGame.ChatTitle))
						answer.ParseMode = "html"
					}
					bot.Send(answer)
				case game.ExtendRegistrationTime:
					game, ok := games[update.Message.Chat.ID]
					if !ok {
						continue
					}

					toStart, err := game.ExtendRegistration(30)

					if err != nil {
						bot.Send(tgbotapi.NewMessage(game.ChatID, fmt.Sprintf("%+v", err)))
					} else {
						bot.Send(tgbotapi.NewMessage(game.ChatID, fmt.Sprintf("Игра начнется через %+v секунд", toStart)))
					}

				}
			case false:
				game, ok := games[update.Message.Chat.ID]
				if !ok {
					continue
				}

				if !game.UserCouldTalk(update.Message.From.UserName) {
					dcfg := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID)
					bot.DeleteMessage(dcfg)
				}
			}
		}
	}
}
