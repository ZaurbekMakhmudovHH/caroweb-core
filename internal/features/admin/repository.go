package admin

import domainuser "carowebapp/core/internal/domain/user"

type Repository interface {
	GetUserByID(userID string) (*User, error)
	SetUserApproved(userID string) error
	SetUserRejected(userID string) error
	InsertUserRejection(userID string, errors map[string]string) error
	ListPendingUsers(search string, limit, offset int) ([]*domainuser.User, error)
	GetUserProfile(userID string) (*domainuser.Profile, error)
}
