package authorization

import (
	"github.com/rredkovich/amazingMafiaBotTG/game"
)

// UserCouldModifyGame returns if given user could do anything with given game
// TODO: do check regarding user's group role (an admin should be able to do smth with the game)
func UserCouldModifyGame(user *game.TGUser, game *game.Game) bool {
	return game.GameInitiator.ID == user.ID
}
