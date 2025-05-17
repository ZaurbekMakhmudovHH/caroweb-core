package auth

import "time"

const (
	RoleHomeowner        = "ROLE_HOMEOWNER"
	RoleManager          = "ROLE_MANAGER"
	RoleAdmin            = "ROLE_ADMIN"
	StatusCreated        = "created"
	StatusEmailConfirmed = "email_confirmed"
	StatusPending        = "pending"
	StatusApproved       = "approved"
	StatusRejected       = "rejected"
)

type User struct {
	ID                     string       `db:"id" json:"id"`
	Email                  string       `db:"email" json:"email"`
	Password               string       `db:"password" json:"-"`
	Role                   string       `db:"role" json:"role"`
	EmailConfirmed         bool         `db:"email_confirmed" json:"email_confirmed"`
	EmailConfirmationToken *string      `db:"email_confirmation_token" json:"email_confirmation_token"`
	LastConfirmationSentAt *time.Time   `db:"last_confirmation_sent_at" json:"last_confirmation_sent_at"`
	Status                 string       `db:"status" json:"status"`
	CreatedAt              time.Time    `db:"created_at" json:"created_at"`
	Profile                *UserProfile `db:"-" json:"profile"`
}

type UserProfile struct {
	UserID      string    `db:"user_id" json:"user_id"`
	Salutation  string    `db:"salutation" json:"salutation"`
	Title       *string   `db:"title" json:"title,omitempty"`
	FirstName   string    `db:"first_name" json:"first_name"`
	LastName    string    `db:"last_name" json:"last_name"`
	Street      string    `db:"street" json:"street"`
	HouseNumber string    `db:"house_number" json:"house_number"`
	PostalCode  string    `db:"postal_code" json:"postal_code"`
	City        string    `db:"city" json:"city"`
	IsVerified  bool      `db:"is_verified" json:"is_verified"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type RefreshToken struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

type UserPasswordResetToken struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	Token     string     `db:"token"`
	CreatedAt time.Time  `db:"created_at"`
	UsedAt    *time.Time `db:"used_at"`
	ExpiresAt time.Time  `db:"expires_at"`
}
