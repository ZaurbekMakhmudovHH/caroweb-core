package auth

import "time"

type Repository interface {
	Create(user *User) (*User, error)
	Update(user *User) error
	GetByConfirmationToken(token string) (*User, error)
	GetByID(id string) (*User, error)

	EmailExists(email string) (bool, error)
	GetByEmail(email string) (*User, error)
	SetEmailConfirmed(userID string) error
	SetUserPending(userID string) error
	UpdateEmailConfirmation(userID string, token string, sentAt time.Time) error

	CreateProfile(profile *UserProfile) error

	StoreRefreshToken(token *RefreshToken) error
	GetRefreshToken(tokenStr string) (*RefreshToken, error)
	DeleteRefreshTokensByUserID(userID string) error

	CreatePasswordResetToken(token *UserPasswordResetToken) error
	GetPasswordResetToken(token string) (*UserPasswordResetToken, error)
	MarkResetTokenUsed(token string) error
	UpdateUserPassword(userID, newHashedPassword string) error
}
