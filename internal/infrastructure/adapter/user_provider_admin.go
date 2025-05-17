package adapter

import (
	"carowebapp/core/internal/domain/user"

	"carowebapp/core/internal/features/admin"

	"context"
)

// AdminUserProvider adapts admin.Repository to the user.Provider interface.
type AdminUserProvider struct {
	Repo admin.Repository
}

// GetByID retrieves a user by ID using admin.Repository and converts it to the domain user model.
func (p *AdminUserProvider) GetByID(_ context.Context, id string) (*user.User, error) {
	raw, err := p.Repo.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}

	return &user.User{
		ID:     raw.ID,
		Email:  raw.Email,
		Status: raw.Status,
		Role:   raw.Role,
	}, nil
}
