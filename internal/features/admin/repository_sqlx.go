package admin

import (
	domainuser "carowebapp/core/internal/domain/user"

	"database/sql"

	"encoding/json"

	"errors"

	"fmt"

	"github.com/jmoiron/sqlx"
)

// sqlxRepository provides SQL-backed implementation of the admin.Repository interface.
type sqlxRepository struct {
	db *sqlx.DB
}

// NewSQLXRepository creates a new instance of sqlxRepository using the given database connection.
func NewSQLXRepository(db *sqlx.DB) Repository {
	return &sqlxRepository{db: db}
}

// User represents a minimal user view used internally by the admin feature.
type User struct {
	ID     string `db:"id"`
	Email  string `db:"email"`
	Status string `db:"status"`
	Role   string `db:"role"`
}

// GetUserByID fetches a user by their ID.
func (r *sqlxRepository) GetUserByID(userID string) (*User, error) {
	var user User
	err := r.db.Get(&user, `
		SELECT id, email, status, role
		FROM users
		WHERE id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// SetUserApproved updates a user's status to 'approved'.
func (r *sqlxRepository) SetUserApproved(userID string) error {
	_, err := r.db.Exec(`
		UPDATE users
		SET status = 'approved'
		WHERE id = $1
	`, userID)
	return err
}

// SetUserRejected updates a user's status to 'rejected'.
func (r *sqlxRepository) SetUserRejected(userID string) error {
	_, err := r.db.Exec(`
		UPDATE users
		SET status = 'rejected'
		WHERE id = $1
	`, userID)
	return err
}

// InsertUserRejection stores the rejection reasons for a given user as a JSON object.
func (r *sqlxRepository) InsertUserRejection(userID string, errors map[string]string) error {
	jsonData, err := json.Marshal(errors)
	if err != nil {
		return fmt.Errorf("failed to marshal rejection errors: %w", err)
	}

	_, err = r.db.Exec(`
		INSERT INTO user_rejections (user_id, errors)
		VALUES ($1, $2)
	`, userID, jsonData)

	return err
}

// ListPendingUsers returns a paginated list of users with 'pending' status,
// optionally filtered by partial match on first or last name.
func (r *sqlxRepository) ListPendingUsers(search string, limit, offset int) ([]*domainuser.User, error) {
	query := `
		SELECT 
			u.id, 
			u.email, 
			u.status,
			COALESCE(p.first_name, '') AS first_name,
			COALESCE(p.last_name, '') AS last_name
		FROM users u
		LEFT JOIN user_profiles p ON u.id = p.user_id
		WHERE u.status = 'pending'
		AND (p.first_name ILIKE '%' || $1 || '%' OR p.last_name ILIKE '%' || $1 || '%')
		ORDER BY u.created_at DESC
		LIMIT $2 OFFSET $3
	`

	type UserWithProfile struct {
		ID        string `db:"id"`
		Email     string `db:"email"`
		Status    string `db:"status"`
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
	}

	var results []UserWithProfile
	err := r.db.Select(&results, query, search, limit, offset)
	if err != nil {
		return nil, err
	}

	var users []*domainuser.User
	for _, res := range results {
		users = append(users, &domainuser.User{
			ID:     res.ID,
			Email:  res.Email,
			Status: res.Status,
			Profile: &domainuser.Profile{
				FirstName: res.FirstName,
				LastName:  res.LastName,
			},
		})
	}

	return users, nil
}

// GetUserProfile retrieves a user's profile information by their user ID.
func (r *sqlxRepository) GetUserProfile(userID string) (*domainuser.Profile, error) {
	var profile domainuser.Profile
	err := r.db.Get(&profile, `
		SELECT user_id, salutation, title, first_name, last_name,
		       street, house_number, postal_code, city, updated_at
		FROM user_profiles
		WHERE user_id = $1
	`, userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &profile, nil
}
