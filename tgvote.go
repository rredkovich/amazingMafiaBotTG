package main

import (
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rredkovich/amazingMafiaBotTG/game"
	"github.com/rredkovich/amazingMafiaBotTG/types"
	"github.com/segmentio/ksuid"
)

type TGVoteValue struct {
	UUID        ksuid.KSUID
	VoteID      string
	GroupChatID int64
	Value       string
	User        *tgbotapi.User
	Avalability game.VoteAvailability
}

func NewTGVoteValue(voteID string, groupChatID int64, value string, av game.VoteAvailability) *TGVoteValue {
	uuid := ksuid.New()
	return &TGVoteValue{
		VoteID:      voteID,
		GroupChatID: groupChatID,
		Value:       value,
		UUID:        uuid,
		Avalability: av,
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

				vote := NewTGVoteValue(v.ID, v.GameChatID, value.Value, v.VoteAvailability)
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
				// TODO operate with userID's everywhere?
				if value.Value == voter.UserName {
					continue
				}
				vote := NewTGVoteValue(v.ID, v.GameChatID, value.Value, game.VoteAvailabilityEnum.Mafia)
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
	case game.VoteAvailabilityEnum.Doctor:

		kbdRows := make([][]tgbotapi.InlineKeyboardButton, 0, len(v.Values)) // will vote for himself
		for _, value := range v.Values {
			// will vote for himself
			// value.Text here has mafia emoji
			// TODO operate with userID's everywhere
			vote := NewTGVoteValue(v.ID, v.GameChatID, value.Value, game.VoteAvailabilityEnum.Doctor)
			voteUUID := vote.UUIDString()
			votes[voteUUID] = vote
			row := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(value.Text, voteUUID)}
			kbdRows = append(kbdRows, row)
		}
		msg := tgbotapi.NewMessage(int64(v.Voters[0].ID), v.VoteText)
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(kbdRows...)
		messages := []*tgbotapi.MessageConfig{&msg}

		return messages

	case game.VoteAvailabilityEnum.Commissar:
		kbdRows := make([][]tgbotapi.InlineKeyboardButton, 0, len(v.Values)-1) // will not vote for himself
		for _, value := range v.Values {
			// will not vote for himself
			if value.Value == v.Voters[0].UserName {
				continue
			}
			// TODO operate with userID's everywhere?
			vote := NewTGVoteValue(v.ID, v.GameChatID, value.Value, game.VoteAvailabilityEnum.Commissar)
			voteUUID := vote.UUIDString()
			votes[voteUUID] = vote
			row := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(value.Text, voteUUID)}
			kbdRows = append(kbdRows, row)
		}
		msg := tgbotapi.NewMessage(int64(v.Voters[0].ID), v.VoteText)
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(kbdRows...)
		messages := []*tgbotapi.MessageConfig{&msg}

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

// EndVote returs result of a vote. nil if noone has voted
func (v *Vote) EndVote() *game.VoteCommandValue {
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
			return value
		}
	}

	return nil
}
