package auth

import (
	"github.com/jmoiron/sqlx"

	"time"
)

type SQLXRepository struct {
	db *sqlx.DB
}

func NewSQLXRepository(db *sqlx.DB) *SQLXRepository {
	return &SQLXRepository{db: db}
}

func (r *SQLXRepository) Create(user *User) (*User, error) {
	user.Status = StatusCreated

	query := `
		INSERT INTO users (
			id, email, password, role, email_confirmation_token,
			last_confirmation_sent_at, email_confirmed, created_at,
		    status
		)
		VALUES (
			:id, :email, :password, :role, :email_confirmation_token,
			:last_confirmation_sent_at, :email_confirmed, :created_at,
			:status
		)
	`

	_, err := r.db.NamedExec(query, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *SQLXRepository) EmailExists(email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`
	err := r.db.Get(&exists, query, email)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *SQLXRepository) GetByEmail(email string) (*User, error) {
	query := `
	SELECT 
		u.id, u.email, u.password, u.role, 
		u.email_confirmed, u.email_confirmation_token, 
		u.last_confirmation_sent_at, u.created_at,
		u.status,
		(up.user_id IS NOT NULL) AS has_profile
	FROM users u
	LEFT JOIN user_profiles up ON u.id = up.user_id
	WHERE u.email = $1
	`

	type dbResult struct {
		User
		HasProfile bool `db:"has_profile"`
	}

	var result dbResult

	if err := r.db.Get(&result, query, email); err != nil {
		return nil, err
	}

	if result.HasProfile {
		result.User.Profile = &UserProfile{
			UserID: result.User.ID,
		}
	}

	return &result.User, nil
}

func (r *SQLXRepository) GetByConfirmationToken(token string) (*User, error) {
	var user User
	err := r.db.Get(&user, `SELECT * FROM users WHERE email_confirmation_token = $1`, token)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *SQLXRepository) SetEmailConfirmed(userID string) error {
	_, err := r.db.Exec(`UPDATE users SET email_confirmed = true, status = $1, email_confirmation_token = NULL WHERE id = $2`, StatusEmailConfirmed, userID)
	return err
}

func (r *SQLXRepository) SetUserPending(userID string) error {
	_, err := r.db.Exec(`UPDATE users SET status = $1 WHERE id = $2`, StatusPending, userID)
	return err
}

func (r *SQLXRepository) UpdateEmailConfirmation(userID string, token string, sentAt time.Time) error {
	_, err := r.db.Exec(`
		UPDATE users
		SET email_confirmation_token = $1, last_confirmation_sent_at = $2
		WHERE id = $3
	`, token, sentAt, userID)
	return err
}

func (r *SQLXRepository) CreateProfile(profile *UserProfile) error {
	query := `
		INSERT INTO user_profiles (
			user_id, salutation, title, first_name, last_name,
			street, house_number, postal_code, city, updated_at
		)
		VALUES (
			:user_id, :salutation, :title, :first_name, :last_name,
			:street, :house_number, :postal_code, :city, :updated_at
		)
	`
	_, err := r.db.NamedExec(query, profile)
	return err
}

func (r *SQLXRepository) GetByID(id string) (*User, error) {
	var user User
	err := r.db.Get(&user, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *SQLXRepository) Update(user *User) error {
	query := `
		UPDATE users
		SET email = :email,
		    password = :password,
		    role = :role,
		    email_confirmation_token = :email_confirmation_token,
		    last_confirmation_sent_at = :last_confirmation_sent_at,
		    email_confirmed = :email_confirmed,
		    created_at = :created_at
		WHERE id = :id
	`
	_, err := r.db.NamedExec(query, user)
	return err
}

func (r *SQLXRepository) StoreRefreshToken(token *RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
		VALUES (:id, :user_id, :token, :expires_at, :created_at)
	`
	_, err := r.db.NamedExec(query, token)
	return err
}

func (r *SQLXRepository) GetRefreshToken(tokenStr string) (*RefreshToken, error) {
	var token RefreshToken
	err := r.db.Get(&token, "SELECT * FROM refresh_tokens WHERE token=$1", tokenStr)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *SQLXRepository) DeleteRefreshTokensByUserID(userID string) error {
	_, err := r.db.Exec("DELETE FROM refresh_tokens WHERE user_id = $1", userID)
	return err
}

func (r *SQLXRepository) CreatePasswordResetToken(token *UserPasswordResetToken) error {
	query := `
		INSERT INTO user_password_reset_tokens (id, user_id, token, created_at, expires_at)
		VALUES (:id, :user_id, :token, :created_at, :expires_at)
	`
	_, err := r.db.NamedExec(query, token)
	return err
}

func (r *SQLXRepository) GetPasswordResetToken(token string) (*UserPasswordResetToken, error) {
	var result UserPasswordResetToken
	query := `
		SELECT id, user_id, token, created_at, used_at, expires_at
		FROM user_password_reset_tokens
		WHERE token = $1
	`
	err := r.db.Get(&result, query, token)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *SQLXRepository) MarkResetTokenUsed(token string) error {
	query := `
		UPDATE user_password_reset_tokens
		SET used_at = NOW()
		WHERE token = $1
	`
	_, err := r.db.Exec(query, token)
	return err
}

func (r *SQLXRepository) UpdateUserPassword(userID, newHashedPassword string) error {
	query := `
		UPDATE users
		SET password = $1
		WHERE id = $2
	`
	_, err := r.db.Exec(query, newHashedPassword, userID)
	return err
}
