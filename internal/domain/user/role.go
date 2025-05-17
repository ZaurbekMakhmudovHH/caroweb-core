package user

const (
	RoleAdmin     = "ROLE_ADMIN"
	RoleManager   = "ROLE_MANAGER"
	RoleHomeowner = "ROLE_HOMEOWNER"
)

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) IsManager() bool {
	return u.Role == RoleManager
}

func (u *User) IsHomeowner() bool {
	return u.Role == RoleHomeowner
}
