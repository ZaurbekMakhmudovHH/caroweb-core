package auth

import (
	"carowebapp/core/internal/infrastructure/email"

	"crypto/rand"

	"encoding/hex"

	"errors"

	"github.com/golang-jwt/jwt/v5"

	"github.com/google/uuid"

	"golang.org/x/crypto/bcrypt"

	"os"

	"strings"

	"time"
)

var (
	ErrEmailExists        = errors.New("email already registered")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrWeakPassword       = errors.New("password too weak")
	ErrInvalidRole        = errors.New("registration with this role is not allowed")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAlreadyConfirmed   = errors.New("email already confirmed")
)

// Service provides authentication and user management functionality.
type Service struct {
	repo   Repository
	Sender email.Sender
}

// NewService creates a new instance of the Service.
func NewService(repo Repository, sender email.Sender) *Service {
	return &Service{
		repo:   repo,
		Sender: sender,
	}
}

// RegisterUser registers a new user with the specified email, password, and role.
// If calledFromScript is true, admin users can be registered.
func (s *Service) RegisterUser(email, password, role string, calledFromScript ...bool) (*User, error) {
	email = strings.TrimSpace(email)

	isCalledFromScript := len(calledFromScript) > 0 && calledFromScript[0]

	if !isCalledFromScript && !isValidRole(role) {
		return nil, ErrInvalidRole
	}

	if len(password) < 6 {
		return nil, ErrWeakPassword
	}

	exists, err := s.repo.EmailExists(email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	token := generateToken()
	now := time.Now()

	user := &User{
		ID:                     uuid.New().String(),
		Email:                  email,
		Password:               string(hashedPassword),
		Role:                   role,
		EmailConfirmed:         false,
		EmailConfirmationToken: &token,
		CreatedAt:              now,
		LastConfirmationSentAt: &now,
	}

	go func() {
		_ = s.Sender.SendConfirmation(user.Email, token)
	}()

	return s.repo.Create(user)
}

// Login authenticates a user with the provided email and password.
// It generates and returns access and refresh tokens upon successful authentication.
func (s *Service) Login(email, password string) (*User, string, string, error) {
	email = strings.TrimSpace(email)
	if email == "" || password == "" {
		return nil, "", "", ErrInvalidCredentials
	}

	user, err := s.repo.GetByEmail(email)
	if err != nil || user == nil {
		return nil, "", "", ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	if err := s.repo.DeleteRefreshTokensByUserID(user.ID); err != nil {
		return nil, "", "", err
	}

	refreshTokenStr := generateToken()
	accessToken, err := generateJWT(user.ID)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken := &RefreshToken{
		ID:        uuid.New().String(),
		Token:     refreshTokenStr,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	if err := s.repo.StoreRefreshToken(refreshToken); err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshTokenStr, nil
}

// GetByConfirmationToken retrieves a user by their email confirmation token.
func (s *Service) GetByConfirmationToken(token string) (*User, error) {
	return s.repo.GetByConfirmationToken(token)
}

// ConfirmEmail confirms a user's email by setting the email_confirmed field to true.
func (s *Service) ConfirmEmail(userID string) error {
	return s.repo.SetEmailConfirmed(userID)
}

// ResendConfirmationByUserID resends an email confirmation link to the user.
func (s *Service) ResendConfirmationByUserID(userID string) error {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}

	if user.EmailConfirmed {
		return ErrAlreadyConfirmed
	}

	if user.LastConfirmationSentAt != nil && time.Since(*user.LastConfirmationSentAt) < time.Minute {
		return errors.New("please wait before requesting another email")
	}

	token := generateToken()
	now := time.Now()

	user.EmailConfirmationToken = &token
	user.LastConfirmationSentAt = &now

	if err := s.repo.Update(user); err != nil {
		return err
	}

	go func() {
		_ = s.Sender.SendConfirmation(user.Email, token)
	}()

	return nil
}

// AddUserProfile creates and associates a user profile with the user.
// The user's status is set to "PENDING" after the profile is added.
func (s *Service) AddUserProfile(profile *UserProfile) error {
	user, err := s.repo.GetByID(profile.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	if user.Profile != nil {
		return errors.New("profile already exists")
	}

	err = s.repo.CreateProfile(profile)
	if err != nil {
		return err
	}

	return s.repo.SetUserPending(profile.UserID)
}

// RequestPasswordReset sends a password reset link to the specified email.
func (s *Service) RequestPasswordReset(email string) error {
	user, err := s.repo.GetByEmail(email)
	if err != nil || user == nil {
		return nil
	}

	token := generateToken()
	now := time.Now()

	reset := &UserPasswordResetToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     token,
		CreatedAt: now,
		ExpiresAt: now.Add(30 * time.Minute),
	}

	if err := s.repo.CreatePasswordResetToken(reset); err != nil {
		return err
	}

	return s.Sender.SendResetPasswordLink(email, token)
}

// IsResetTokenValid checks if the given password reset token is still valid.
func (s *Service) IsResetTokenValid(token string) (bool, error) {
	reset, err := s.repo.GetPasswordResetToken(token)
	if err != nil || reset == nil {
		return true, err
	}

	if reset.UsedAt != nil || time.Now().After(reset.ExpiresAt) {
		return false, nil
	}

	return true, nil
}

// ResetPassword updates the user's password using the provided reset token.
func (s *Service) ResetPassword(token string, newPassword string) error {
	reset, err := s.repo.GetPasswordResetToken(token)
	if err != nil || reset == nil {
		return errors.New("invalid or expired token")
	}

	if reset.UsedAt != nil || time.Now().After(reset.ExpiresAt) {
		return errors.New("token is no longer valid")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateUserPassword(reset.UserID, string(hashed)); err != nil {
		return err
	}

	return s.repo.MarkResetTokenUsed(token)
}

// generateToken generates a secure random token as a string.
func isValidRole(role string) bool {
	switch role {
	case RoleHomeowner, RoleManager:
		return true
	default:
		return false
	}
}

// generateJWT generates a JSON Web Token (JWT) for the specified user ID.
func generateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// isValidRole checks if the provided role is one of the allowed roles.
func generateJWT(userID string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
