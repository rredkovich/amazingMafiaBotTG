package authorization

import (
	"github.com/rredkovich/amazingMafiaBotTG/types"
	"testing"
)

func TestUserCouldModifyGame(t *testing.T) {
	userID := 42
	user := types.TGUser{userID, "", "", "sanddog", "ru", false}
	game := types.Game{
		ChatID:        10,
		GameInitiator: user,
	}

	couldDO := UserCouldModifyGame(&user, &game)

	if !couldDO {
		t.Errorf("Expected user cannot do changes to a game")
	}

	user.ID = 10

	couldDO = UserCouldModifyGame(&user, &game)

	if couldDO {
		t.Errorf("Not Expected user can do changes to a game")
	}

}
