package game

import (
	"fmt"
	"github.com/rredkovich/amazingMafiaBotTG/types"
	"testing"
)

var u0 = types.TGUser{
	ID:        1,
	UserName:  "U1",
	FirstName: "User",
	LastName:  "Zero",
	Role:      "",
}

var u1 = types.TGUser{
	ID:        2,
	UserName:  "U2",
	FirstName: "User",
	LastName:  "One",
	Role:      "",
}
var u2 = types.TGUser{
	ID:        3,
	UserName:  "U3",
	FirstName: "User",
	LastName:  "Two",
	Role:      "",
}

var u3 = types.TGUser{
	ID:        4,
	UserName:  "U4",
	FirstName: "User",
	LastName:  "Fourth",
	Role:      "",
}

var u4 = types.TGUser{
	ID:        5,
	UserName:  "U5",
	FirstName: "User",
	LastName:  "Fifth",
	Role:      "",
}

var u5 = types.TGUser{
	ID:        6,
	UserName:  "U6",
	FirstName: "User",
	LastName:  "Sixth",
	Role:      "",
}

func TestGame_AddMember(t *testing.T) {
	ch := make(chan GameMessage)
	vch := make(chan *VoteCommand)

	// AddMember doesn't allow to add member if game is not Prepairing
	g := NewGame(0, "Three users", &u0, &ch, vch)
	g.State.SetStarted()
	err := g.AddMember(&u1)

	if err == nil {
		t.Errorf("No error on attempt to add user to game in progress, err: %+v\n", err)
	}

	// AddMember assigns peace role for all added members
	g = NewGame(0, "Three users", &u0, &ch, vch)
	g.State.SetPrepairing()
	g.AddMember(&u1)
	g.AddMember(&u2)
	g.AddMember(&u3)

	for _, member := range g.Members {
		if member.Role != types.Citizen {
			t.Errorf("Added member %+v has role '%+v', but shoud has '%+v'\n", member, member.Role, types.Citizen)
		}
	}
}

func TestGame_assignBaseRole(t *testing.T) {
	ch := make(chan GameMessage)
	vch := make(chan *VoteCommand)
	var u = types.TGUser{
		ID:        1,
		UserName:  "U1",
		FirstName: "User",
		LastName:  "Zero",
		Role:      "",
	}

	// AddMember doesn't allow to add member if game is not Prepairing
	g := NewGame(0, "Three users", &u0, &ch, vch)

	g.assignBaseRole(&u)

	if u.Role != types.Citizen {
		t.Errorf("Base role hasn't been set to user %+v\n", u)
	}

}

func TestGame_AssignRoles_FourUsers(t *testing.T) {

	ch := make(chan GameMessage)
	vch := make(chan *VoteCommand)
	g := NewGame(0, "Three users", &u0, &ch, vch)
	g.State.SetPrepairing()
	g.AddMember(&u1)
	g.AddMember(&u2)
	g.AddMember(&u3)

	g.AssignRoles()

	// No commisar for game less than 4 users
	if g.Commissar != nil {
		t.Errorf("Commissar '%+v' was set for game of %+v members", g.Commissar, len(g.Members))
	}
	if g.Doctor == nil {
		t.Errorf("Doctor '%+v' was not set", g.Doctor)
	}

	if g.Doctor == g.Commissar {
		t.Errorf("Doctor and commissar is same user")
	}

	if len(g.GangsterMembers) != 1 {
		t.Errorf("%+v gangsters were chosen but should be exactly one in case of three users", len(g.GangsterMembers))
	}

	_, docIsGangsta := g.GangsterMembers[g.Doctor.UserName]
	if docIsGangsta {
		t.Errorf("Doctor was chosen as a gangster")
	}
}

func TestGame_AssignRoles_SixUsers(t *testing.T) {
	ch := make(chan GameMessage)
	vch := make(chan *VoteCommand)
	g := NewGame(0, "Six users", &u0, &ch, vch)
	g.State.SetPrepairing()
	g.AddMember(&u1)
	g.AddMember(&u2)
	g.AddMember(&u3)
	g.AddMember(&u4)
	g.AddMember(&u5)

	g.AssignRoles()

	if g.Doctor == nil || g.Commissar == nil {
		t.Errorf("Doctor '%+v' or commissar '%+v' were not set", g.Doctor, g.Commissar)
	}

	if g.Doctor == g.Commissar {
		t.Errorf("Doctor and commissar is same user")
	}

	numOfGangsters := 2
	if len(g.GangsterMembers) != numOfGangsters {
		t.Errorf("%+v gangsters were chosen but should be exactly %+v in case of six users", len(g.GangsterMembers), numOfGangsters)
	}

	_, docIsGangsta := g.GangsterMembers[g.Doctor.UserName]
	if docIsGangsta {
		t.Errorf("Doctor was chosen as a gangster")
	}

	_, comIsGangsta := g.GangsterMembers[g.Commissar.UserName]
	if comIsGangsta {
		t.Errorf("Commissar was chosen as a gangster")
	}
}

func TestGame_SpecRoleIsDead(t *testing.T) {
	u0 := types.TGUser{
		ID:        1,
		UserName:  "U1",
		FirstName: "User",
		LastName:  "Zero",
		Role:      "",
	}

	mch := make(chan GameMessage)
	vch := make(chan *VoteCommand)

	g := NewGame(100, "test", &u0, &mch, vch)
	g.Commissar = nil
	g.Doctor = nil

	comDead := g.CommissarIsDead()
	docDead := g.DoctorIsDead()

	if !comDead {
		t.Errorf("Commissar is not dead when g.Commisar == nil")
	}

	if !docDead {
		fmt.Printf("docDead %v", docDead)
		t.Errorf("Doctor is not dead when g.Doctor == nil")

	}
}

func TestNewGame(t *testing.T) {
	mch := make(chan GameMessage)
	vch := make(chan *VoteCommand)
	g := NewGame(42, "test", &u0, &mch, vch)
	if g.GameInitiator.UserName != u0.UserName {
		t.Errorf("The Game Initiator was not added as a game member")
	}
}

func TestGame_UserInGame(t *testing.T) {
	mch := make(chan GameMessage)
	vch := make(chan *VoteCommand)
	g := NewGame(42, "test", &u0, &mch, vch)
	_ = g.AddMember(&u1)
	_ = g.AddMember(&u2)

	tests := []struct {
		name string
		game *Game
		user *types.TGUser
		want bool
	}{
		{"In game", g, &u1, true},
		{"In game too", g, &u2, true},
		{"Not in game", g, &u3, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := g.UserInGame(tt.user); got != tt.want {
				t.Errorf("UserInGame() = %v, want %v", got, tt.want)
			}
		})
	}
}
