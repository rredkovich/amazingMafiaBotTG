package game

import (
	"github.com/rredkovich/amazingMafiaBotTG/types"
	"sync"
)

type VoteAction int
type voteActionEnum struct {
	Start VoteAction
	Stop  VoteAction
}

var VoteActionEnum = voteActionEnum{
	0,
	1,
}

const (
	StartVoteAction VoteAction = iota
	StopVoteAction
)

type VoteAvailability int

type voteAvailabilityEnum = struct {
	Lynch     VoteAvailability
	Mafia     VoteAvailability
	Doctor    VoteAvailability
	Commissar VoteAvailability
}

var VoteAvailabilityEnum = voteAvailabilityEnum{0, 1, 2, 3}

type VoteCommand struct {
	sync.Mutex
	GameChatID       int64
	Action           VoteAction
	VoteAvailability VoteAvailability
	VoteText         string
	Voters           []*types.TGUser
	Values           []*VoteCommandValue
	result           *VoteCommandValue
}

func (vc *VoteCommand) SetResult(v *VoteCommandValue) {
	vc.Lock()
	defer vc.Unlock()

	vc.result = v
}

func (vc *VoteCommand) GetResult() *VoteCommandValue {
	vc.Lock()
	defer vc.Unlock()

	return vc.result
}

type VoteCommandValue struct {
	Text  string
	Value string
}
