package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rredkovich/amazingMafiaBotTG/types"
	"log"
	"os"
)

func main() {

	games := make(map[int64]*types.Game)

	//ticker := time.NewTicker(time.Second)
	//done := make(chan bool)

	//go func() {
	//	for {
	//		select {
	//		case <-done:
	//			return
	//		case t := <-ticker.C:
	//			//fmt.Println("Tick at", t)
	//			fmt.Printf("%+v, %v\n", games, t)
	//			for _, game := range games {
	//				// Dirty hack, how it will scale in case of hundred games
	//				go game.Play()
	//			}
	//		}
	//	}
	//}()

	messagesFromGames := make(chan types.GameMessage)

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
			if update.Message.Text == types.LaunchNewGame {
				from := update.Message.From
				starter := types.NewTGUser(from.ID, from.UserName, from.FirstName, from.LastName)
				game := types.NewGame(update.Message.Chat.ID, starter, &messagesFromGames)
				games[game.ChatID] = game
				go game.Play()
				log.Printf("Created game: %+v\n", game)
			} else if update.Message.Text == types.EndGame {
				game, ok := games[update.Message.Chat.ID]
				if !ok {
					continue
				}
				delete(games, game.ChatID)
				game.Stop()
				log.Printf("Stopped game: %+v\n", game)

				tgMsg := tgbotapi.NewMessage(game.ChatID, "The game has been stopped")

				_, err = bot.Send(tgMsg)

				if err != nil {
					log.Printf("Got error on send! %+v\n", err)
				}
			}
		}
	}
}
