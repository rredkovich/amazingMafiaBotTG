package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rredkovich/amazingMafiaBotTG/authorization"
	"github.com/rredkovich/amazingMafiaBotTG/types"
	"log"
	"os"
)

func main() {

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

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		if update.Message.Text == "/start" {
			starter := types.TGUser(*update.Message.From)
			//game := types.Game{update.Message.Chat.ID, starter, nil}
			game := types.NewGame(update.Message.Chat.ID, &starter)
			log.Printf("%+v\n", game)
			log.Printf("%+v\n", game.GameInitiator)
			fmt.Printf("User %v could modify game: %v\n", starter.FirstName, authorization.UserCouldModifyGame(&starter, game))
			fmt.Printf("Starter user is in game: %v\n", game.UserInGame(&starter))
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("<b>Got</b> '%v'", update.Message.Text))
		msg.ReplyToMessageID = update.Message.MessageID

		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("Got error on send! %+v\n", err)
		}
	}
}
