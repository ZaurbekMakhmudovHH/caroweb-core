package user

import "time"

type User struct {
	ID             string
	Email          string
	Role           string
	EmailConfirmed bool
	Status         string
	Profile        *Profile
}

type Profile struct {
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
