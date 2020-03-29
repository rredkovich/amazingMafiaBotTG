package types

import "fmt"

type TGUser struct {
	ID              int
	UserName        string
	FirstName       string
	LastName        string
	Role            RoleType
	SpokenLastWords bool
}

func NewTGUser(id int, userName string, firstName string, lastName string) *TGUser {
	return &TGUser{
		ID:              id,
		UserName:        userName,
		FirstName:       firstName,
		LastName:        lastName,
		SpokenLastWords: false,
	}
}

// InlineTGName returns markdowned type of name wich telegram translates into inline clickable name
func (u *TGUser) InlineTGName() string {
	if u.FirstName != "" && u.LastName != "" {
		return fmt.Sprintf("<a href=\"tg://user?id=%+v\">%+v %+v</a>", u.ID, u.FirstName, u.LastName)
	} else {
		return fmt.Sprintf("<a href=\"tg://user?id=%+v\">%+v</a>", u.ID, u.FirstName)
	}
}

// HumanReadableName returns nice human readable name :shrug:
func (u *TGUser) HumanReadableName() string {
	if u.FirstName != "" && u.LastName != "" {
		return fmt.Sprintf("%+v %+v", u.FirstName, u.LastName)
	} else {
		return fmt.Sprintf("%+v", u.FirstName)
	}

}
