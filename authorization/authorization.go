package authorization

import "github.com/rredkovich/amazingMafiaBotTG/types"

// UserCouldModifyGame returns if given user could do anything with given game
// TODO: do check regarding user's group role (an admin should be able to do smth with the game)
func UserCouldModifyGame(user *types.TGUser, game *types.Game) bool {
	return game.GameInitiator.ID == user.ID
}
