package types

type TGUser struct {
	ID        int
	UserName  string
	FirstName string
	LastName  string
	Role      RoleType
}

func NewTGUser(id int, userName string, firstName string, lastName string) *TGUser {
	return &TGUser{
		ID:        id,
		UserName:  userName,
		FirstName: firstName,
		LastName:  lastName,
	}
}
