package user

import "context"

type Provider interface {
	GetByID(ctx context.Context, id string) (*User, error)
}
