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

	messagesFromGames := make(chan game.GameMessage)

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
		case msg := <-messagesFromGames:
			log.Printf("Games: %+v\n", games)
			fmt.Printf("Got GameMessage %+v\n", msg)
			tgMsg := tgbotapi.NewMessage(msg.ChatID, msg.Message)
			//msg.ReplyToMessageID = update.Message.MessageID

			_, err = bot.Send(tgMsg)

			if err != nil {
				log.Printf("Got error on send! %+v\n", err)
			}
		case update := <-updatesCh:
			log.Printf("Games: %+v\n", games)
			if update.Message == nil {
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
					game := game.NewGame(update.Message.Chat.ID, update.Message.Chat.Title, starter, &messagesFromGames)
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

				if !game.UserCouldTalk(update.Message.From.ID) {
					dcfg := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID)
					bot.DeleteMessage(dcfg)
				}
			}
		}
	}
}
