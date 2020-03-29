package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rredkovich/amazingMafiaBotTG/game"
	"github.com/rredkovich/amazingMafiaBotTG/types"
	"github.com/segmentio/ksuid"
	"reflect"
	"testing"
)

func TestNewTGVoteValue(t *testing.T) {
	type args struct {
		voteID      string
		groupChatID int64
		value       string
		av          game.VoteAvailability
	}
	tests := []struct {
		name string
		args args
		want *TGVoteValue
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTGVoteValue(tt.args.voteID, tt.args.groupChatID, tt.args.value, tt.args.av); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTGVoteValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewVoteFromVoteCommand(t *testing.T) {
	type args struct {
		vcmd    *game.VoteCommand
		ID      string
		botLink string
	}
	tests := []struct {
		name string
		args args
		want *Vote
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewVoteFromVoteCommand(tt.args.vcmd, tt.args.ID, tt.args.botLink); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewVoteFromVoteCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTGVoteValue_UUIDString(t *testing.T) {
	type fields struct {
		UUID         ksuid.KSUID
		VoteID       string
		GroupChatID  int64
		Value        string
		User         *tgbotapi.User
		Availability game.VoteAvailability
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &TGVoteValue{
				UUID:         tt.fields.UUID,
				VoteID:       tt.fields.VoteID,
				GroupChatID:  tt.fields.GroupChatID,
				Value:        tt.fields.Value,
				User:         tt.fields.User,
				Availability: tt.fields.Availability,
			}
			if got := v.UUIDString(); got != tt.want {
				t.Errorf("UUIDString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVote_EndVote(t *testing.T) {
	voters := []*types.TGUser{
		{ID: 1, UserName: "U1"},
		{ID: 2, UserName: "U2"},
		{ID: 3, UserName: "U3"},
	}
	voter4 := &types.TGUser{ID: 4, UserName: "U4"}
	values := []*game.VoteCommandValue{
		{Text: "opt1", Value: "val1"},
		{Text: "opt2", Value: "val2"},
	}
	type fields struct {
		ID               string
		GameChatID       int64
		VoteText         string
		Action           game.VoteAction
		VoteAvailability game.VoteAvailability
		Voters           []*types.TGUser
		Values           []*game.VoteCommandValue
		Votes            map[*tgbotapi.User]string
		botLink          string
	}
	tests := []struct {
		name   string
		fields fields
		want   *game.VoteCommandValue
	}{
		{
			name: "Nobody voted",
			fields: fields{
				ID:               "a",
				GameChatID:       42,
				VoteText:         "do vote",
				Action:           game.VoteActionEnum.Start,
				VoteAvailability: game.VoteAvailabilityEnum.Lynch,
				Voters:           voters,
				Values:           values,
				Votes:            map[*tgbotapi.User]string{},
				botLink:          "",
			},
			want: nil,
		},
		{
			name: "Failed lynch vote (<1/2 voted)",
			fields: fields{
				ID:               "a",
				GameChatID:       42,
				VoteText:         "do vote",
				Action:           game.VoteActionEnum.Start,
				VoteAvailability: game.VoteAvailabilityEnum.Lynch,
				Voters:           voters,
				Values:           values,
				Votes:            map[*tgbotapi.User]string{BotApiUserFromTGUser(voters[0]): "val1"},
				botLink:          "",
			},
			want: nil,
		},
		{
			name: "Succeeded lynch vote (1/2 voted)",
			fields: fields{
				ID:               "a",
				GameChatID:       42,
				VoteText:         "do vote",
				Action:           game.VoteActionEnum.Start,
				VoteAvailability: game.VoteAvailabilityEnum.Lynch,
				Voters:           append(voters, voter4),
				Values:           values,
				Votes: map[*tgbotapi.User]string{
					BotApiUserFromTGUser(voters[0]): "val1",
					BotApiUserFromTGUser(voters[1]): "val1",
				},
				botLink: "",
			},
			want: values[0],
		},
		{
			name: "Unsuccessful lynch vote (1/2 voted, cannot decide)",
			fields: fields{
				ID:               "a",
				GameChatID:       42,
				VoteText:         "do vote",
				Action:           game.VoteActionEnum.Start,
				VoteAvailability: game.VoteAvailabilityEnum.Lynch,
				Voters:           append(voters, voter4),
				Values:           values,
				Votes: map[*tgbotapi.User]string{
					BotApiUserFromTGUser(voters[0]): "val1",
					BotApiUserFromTGUser(voters[1]): "val2",
				},
				botLink: "",
			},
			want: nil,
		},
		{
			name: "Succeeded lynch vote (>1/2 voted)",
			fields: fields{
				ID:               "a",
				GameChatID:       42,
				VoteText:         "do vote",
				Action:           game.VoteActionEnum.Start,
				VoteAvailability: game.VoteAvailabilityEnum.Lynch,
				Voters:           append(voters, voter4),
				Values:           values,
				Votes: map[*tgbotapi.User]string{
					BotApiUserFromTGUser(voters[0]): "val1",
					BotApiUserFromTGUser(voters[1]): "val1",
					BotApiUserFromTGUser(voters[2]): "val1",
				},
				botLink: "",
			},
			want: values[0],
		},
		{
			name: "Succeeded lynch vote (>1/2 voted, majority decided)",
			fields: fields{
				ID:               "a",
				GameChatID:       42,
				VoteText:         "do vote",
				Action:           game.VoteActionEnum.Start,
				VoteAvailability: game.VoteAvailabilityEnum.Lynch,
				Voters:           append(voters, voter4),
				Values:           values,
				Votes: map[*tgbotapi.User]string{
					BotApiUserFromTGUser(voters[0]): "val2",
					BotApiUserFromTGUser(voters[1]): "val2",
					BotApiUserFromTGUser(voters[2]): "val1",
				},
				botLink: "",
			},
			want: values[1],
		},
		{
			name: "Succeeded lynch vote majority decided",
			fields: fields{
				ID:               "a",
				GameChatID:       42,
				VoteText:         "do vote",
				Action:           game.VoteActionEnum.Start,
				VoteAvailability: game.VoteAvailabilityEnum.Lynch,
				Voters:           append(voters, voter4),
				Values:           values,
				Votes: map[*tgbotapi.User]string{
					BotApiUserFromTGUser(voters[0]): "val1",
					BotApiUserFromTGUser(voters[1]): "val1",
					BotApiUserFromTGUser(voters[2]): "val1",
					BotApiUserFromTGUser(voter4):    "val2",
				},
				botLink: "",
			},
			want: values[0],
		},
		{
			name: "Not successful lynch vote majority could not decide",
			fields: fields{
				ID:               "a",
				GameChatID:       42,
				VoteText:         "do vote",
				Action:           game.VoteActionEnum.Start,
				VoteAvailability: game.VoteAvailabilityEnum.Lynch,
				Voters:           append(voters, voter4),
				Values:           values,
				Votes: map[*tgbotapi.User]string{
					BotApiUserFromTGUser(voters[0]): "val1",
					BotApiUserFromTGUser(voters[1]): "val1",
					BotApiUserFromTGUser(voters[2]): "val2",
					BotApiUserFromTGUser(voter4):    "val2",
				},
				botLink: "",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Vote{
				ID:               tt.fields.ID,
				GameChatID:       tt.fields.GameChatID,
				VoteText:         tt.fields.VoteText,
				Action:           tt.fields.Action,
				VoteAvailability: tt.fields.VoteAvailability,
				Voters:           tt.fields.Voters,
				Values:           tt.fields.Values,
				Votes:            tt.fields.Votes,
				botLink:          tt.fields.botLink,
			}
			if got := v.EndVote(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EndVote() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVote_EveryBodyVoted(t *testing.T) {
	type fields struct {
		ID               string
		GameChatID       int64
		VoteText         string
		Action           game.VoteAction
		VoteAvailability game.VoteAvailability
		Voters           []*types.TGUser
		Values           []*game.VoteCommandValue
		Votes            map[*tgbotapi.User]string
		botLink          string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Vote{
				ID:               tt.fields.ID,
				GameChatID:       tt.fields.GameChatID,
				VoteText:         tt.fields.VoteText,
				Action:           tt.fields.Action,
				VoteAvailability: tt.fields.VoteAvailability,
				Voters:           tt.fields.Voters,
				Values:           tt.fields.Values,
				Votes:            tt.fields.Votes,
				botLink:          tt.fields.botLink,
			}
			if got := v.EveryBodyVoted(); got != tt.want {
				t.Errorf("EveryBodyVoted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVote_RegisterVote(t *testing.T) {
	type fields struct {
		ID               string
		GameChatID       int64
		VoteText         string
		Action           game.VoteAction
		VoteAvailability game.VoteAvailability
		Voters           []*types.TGUser
		Values           []*game.VoteCommandValue
		Votes            map[*tgbotapi.User]string
		botLink          string
	}
	type args struct {
		u    *tgbotapi.User
		vote string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Vote{
				ID:               tt.fields.ID,
				GameChatID:       tt.fields.GameChatID,
				VoteText:         tt.fields.VoteText,
				Action:           tt.fields.Action,
				VoteAvailability: tt.fields.VoteAvailability,
				Voters:           tt.fields.Voters,
				Values:           tt.fields.Values,
				Votes:            tt.fields.Votes,
				botLink:          tt.fields.botLink,
			}
			if err := v.RegisterVote(tt.args.u, tt.args.vote); (err != nil) != tt.wantErr {
				t.Errorf("RegisterVote() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVote_StartVote(t *testing.T) {
	type fields struct {
		ID               string
		GameChatID       int64
		VoteText         string
		Action           game.VoteAction
		VoteAvailability game.VoteAvailability
		Voters           []*types.TGUser
		Values           []*game.VoteCommandValue
		Votes            map[*tgbotapi.User]string
		botLink          string
	}
	type args struct {
		votes map[string]*TGVoteValue
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*tgbotapi.MessageConfig
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Vote{
				ID:               tt.fields.ID,
				GameChatID:       tt.fields.GameChatID,
				VoteText:         tt.fields.VoteText,
				Action:           tt.fields.Action,
				VoteAvailability: tt.fields.VoteAvailability,
				Voters:           tt.fields.Voters,
				Values:           tt.fields.Values,
				Votes:            tt.fields.Votes,
				botLink:          tt.fields.botLink,
			}
			if got := v.StartVote(tt.args.votes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StartVote() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BotApiUserFromTGUser(u *types.TGUser) *tgbotapi.User {
	return &tgbotapi.User{
		ID:           1,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		UserName:     u.UserName,
		LanguageCode: "ru",
		IsBot:        false,
	}
}
