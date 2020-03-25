package main

import (
	"encoding/base64"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rredkovich/amazingMafiaBotTG/game"
	"github.com/rredkovich/amazingMafiaBotTG/types"
	"log"
	"os"
	"strconv"
)

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

	bot.Debug = false // Has the library display every request and response.

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 10

	updatesCh, err := bot.GetUpdatesChan(u)

	for {
		select {
		case cmd := <-voteCommandsFromGames:
			voteKey := fmt.Sprintf("%+v_%v", cmd.GameChatID, cmd.VoteAvailability)
			fmt.Printf("Got vote command %+v\n", cmd)
			switch cmd.Action {
			case game.StartVoteAction:
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
				result := vote.EndVote()
				log.Printf("Setting result '%+v' for vote %+v:'%+v'", result, voteKey, vote.VoteText)
				//if err != nil {
				//	fmt.Printf("Vote step into shit %+v", err)
				//	delete(votes, voteKey)
				//	continue
				//}
				cmd.SetResult(result, true)
				delete(votes, voteKey)
			}

		case msg := <-messagesFromGames:
			//log.Printf("Games: %+v\n", games)
			//log.Printf("Got GameMessage %+v\n", msg)
			tgMsg := tgbotapi.NewMessage(msg.ChatID, msg.Message)
			tgMsg.ParseMode = "html"

			_, err = bot.Send(tgMsg)

			if err != nil {
				log.Printf("Got error on send! %+v\n", err)
			}

			// handling case when game for chat has been stopped and we received last message with results
			game, ok := games[msg.ChatID]
			if !ok {
				continue
			}

			if game.State.IsStopped() {
				delete(games, msg.ChatID)
			}

		case update := <-updatesCh:
			log.Printf("Games: %+v\n", games)
			// command execution from voting
			if update.Message == nil {
				//log.Printf("VOTE ENUM %s", game.VoteActionEnum.Start)
				//log.Printf("%+v\n", update.CallbackQuery)
				tgVote, ok := tgVotes[update.CallbackQuery.Data]
				if !ok {
					fmt.Printf("Cannot get vote by callback query '%+v'", update.CallbackQuery.Data)
					continue
				}
				vote := votes[tgVote.VoteID]
				delete(tgVotes, update.CallbackQuery.Data)

				var answerText string
				err := vote.RegisterVote(update.CallbackQuery.From, tgVote.Value)
				if err != nil {
					answerText = fmt.Sprintf("%+v", err)
				} else {
					answerText = fmt.Sprintf("Вы выбрали %+v", vote.)

				}
				bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "👌"))
				bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, answerText))

				switch vote.VoteAvailability {
				case game.VoteAvailabilityEnum.Mafia:
					msg := tgbotapi.NewMessage(vote.GameChatID, "<b>Мафия</b> выбрала жертву")
					msg.ParseMode = "html"
					bot.Send(msg)
				case game.VoteAvailabilityEnum.Doctor:
					msg := tgbotapi.NewMessage(vote.GameChatID, "<b>Доктор</b> достал бинты и мазь")
					msg.ParseMode = "html"
					bot.Send(msg)
				case game.VoteAvailabilityEnum.Commissar:
					msg := tgbotapi.NewMessage(vote.GameChatID, "<b>Комиссар Каттани</b> притворился петуньей в горшке у дома подозреваемого...")
					msg.ParseMode = "html"
					bot.Send(msg)
				case game.VoteAvailabilityEnum.Lynch:
					msg := tgbotapi.NewMessage(vote.GameChatID, fmt.Sprintf("@%+v проголосовал", update.CallbackQuery.From.UserName))
					msg.ParseMode = "html"
					bot.Send(msg)
				}
				toDelete := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
				bot.DeleteMessage(toDelete)
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
