package game

import (
	"fmt"
	"github.com/rredkovich/amazingMafiaBotTG/types"
	"testing"
)

func TestGame_AssignRoles_ThreeUsers(t *testing.T) {
	u0 := types.TGUser{
		ID:        1,
		UserName:  "U1",
		FirstName: "User",
		LastName:  "Zero",
		Role:      "",
	}
	u1 := types.TGUser{
		ID:        2,
		UserName:  "U2",
		FirstName: "User",
		LastName:  "One",
		Role:      "",
	}
	u2 := types.TGUser{
		ID:        3,
		UserName:  "U3",
		FirstName: "User",
		LastName:  "Two",
		Role:      "",
	}

	ch := make(chan GameMessage)
	vch := make(chan *VoteCommand)
	g := NewGame(0, "Three users", &u0, &ch, vch)
	g.State.SetPrepairing()
	g.AddMember(&u1)
	g.AddMember(&u2)

	g.AssignRoles()

	if g.Doctor == nil || g.Commissar == nil {
		t.Errorf("Doctor '%+v' or commissar '%+v' were not set", g.Doctor, g.Commissar)
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

	_, comIsGangsta := g.GangsterMembers[g.Commissar.UserName]
	if comIsGangsta {
		t.Errorf("Commissar was chosen as a gangster")
	}
}

func TestGame_AssignRoles_SixUsers(t *testing.T) {
	u0 := types.TGUser{
		ID:        1,
		UserName:  "U1",
		FirstName: "User",
		LastName:  "Zero",
		Role:      "",
	}
	u1 := types.TGUser{
		ID:        2,
		UserName:  "U2",
		FirstName: "User",
		LastName:  "One",
		Role:      "",
	}
	u2 := types.TGUser{
		ID:        3,
		UserName:  "U3",
		FirstName: "User",
		LastName:  "Two",
		Role:      "",
	}
	u3 := types.TGUser{
		ID:        4,
		UserName:  "U4",
		FirstName: "User",
		LastName:  "Fourth",
		Role:      "",
	}
	u4 := types.TGUser{
		ID:        5,
		UserName:  "U5",
		FirstName: "User",
		LastName:  "Fifth",
		Role:      "",
	}
	u5 := types.TGUser{
		ID:        6,
		UserName:  "U6",
		FirstName: "User",
		LastName:  "Sixth",
		Role:      "",
	}

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

	if len(g.GangsterMembers) != 2 {
		t.Errorf("%+v gangsters were chosen but should be exactly one in case of three users", len(g.GangsterMembers))
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
