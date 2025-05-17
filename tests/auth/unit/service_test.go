package unit

import (
	"carowebapp/core/internal/features/auth"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"

	"testing"

	"time"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) GetByID(id string) (*auth.User, error) {
	args := m.Called(id)
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockUserRepo) Update(user *auth.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepo) UpdateEmailConfirmation(userID string, token string, sentAt time.Time) error {
	args := m.Called(userID, token, sentAt)
	return args.Error(0)
}

func (m *MockUserRepo) CreateProfile(profile *auth.UserProfile) error {
	args := m.Called(profile)
	return args.Error(0)
}

func (m *MockUserRepo) StoreRefreshToken(token *auth.RefreshToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockUserRepo) GetRefreshToken(tokenStr string) (*auth.RefreshToken, error) {
	args := m.Called(tokenStr)
	return args.Get(0).(*auth.RefreshToken), args.Error(1)
}

func (m *MockUserRepo) DeleteRefreshTokensByUserID(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockUserRepo) Create(user *auth.User) (*auth.User, error) {
	args := m.Called(user)
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockUserRepo) GetByEmail(email string) (*auth.User, error) {
	args := m.Called(email)
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockUserRepo) EmailExists(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepo) GetByConfirmationToken(token string) (*auth.User, error) {
	args := m.Called(token)
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockUserRepo) SetEmailConfirmed(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockUserRepo) CreatePasswordResetToken(token *auth.UserPasswordResetToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockUserRepo) GetPasswordResetToken(token string) (*auth.UserPasswordResetToken, error) {
	args := m.Called(token)
	return args.Get(0).(*auth.UserPasswordResetToken), args.Error(1)
}

func (m *MockUserRepo) UpdateUserPassword(userID string, hashed string) error {
	args := m.Called(userID, hashed)
	return args.Error(0)
}

func (m *MockUserRepo) MarkResetTokenUsed(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

type MockSender struct {
	mock.Mock
}

func (m *MockSender) SendMail(to string, subject string, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

func (m *MockSender) SendConfirmation(to string, token string) error {
	args := m.Called(to, token)
	return args.Error(0)
}

func (m *MockSender) SendApprovalNotification(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *MockSender) SendRejectionNotification(email string, errors map[string]string) error {
	args := m.Called(email, errors)
	return args.Error(0)
}

func (m *MockSender) SendResetPasswordLink(to string, token string) error {
	args := m.Called(to, token)
	return args.Error(0)
}

// TestRegisterUser_Success verifies that a new user can be registered successfully
// and that a confirmation email is sent.
func TestRegisterUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepo)
	mockMailer := new(MockSender)

	svc := auth.NewService(mockRepo, mockMailer)

	email := "test@example.com"
	password := "securepass"
	role := "ROLE_HOMEOWNER"

	expectedUser := &auth.User{
		Email: email,
		Role:  role,
	}

	mockRepo.On("Create", mock.AnythingOfType("*auth.User")).Return(expectedUser, nil)
	mockMailer.On("SendConfirmation", email, mock.AnythingOfType("string")).Return(nil)

	user, err := svc.RegisterUser(email, password, role)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser.Email, user.Email)
	assert.Equal(t, expectedUser.Role, user.Role)

	mockRepo.AssertExpectations(t)
	mockMailer.AssertExpectations(t)
}

// TestRegisterUser_EmailAlreadyExists verifies that registration fails
// when the email already exists in the system.
func TestRegisterUser_EmailAlreadyExists(t *testing.T) {
	mockRepo := new(MockUserRepo)
	mockMailer := new(MockSender)
	svc := auth.NewService(mockRepo, mockMailer)

	email := "existing@example.com"
	password := "securepass"
	role := "ROLE_HOMEOWNER"

	mockRepo.On("EmailExists", email).Return(true, nil)

	user, err := svc.RegisterUser(email, password, role)

	assert.Nil(t, user)
	assert.ErrorIs(t, err, auth.ErrEmailExists)

	mockRepo.AssertExpectations(t)
}

// TestRegisterUser_WeakPassword verifies that registration fails
// when the provided password does not meet strength requirements.
func TestRegisterUser_WeakPassword(t *testing.T) {
	mockRepo := new(MockUserRepo)
	mockMailer := new(MockSender)
	svc := auth.NewService(mockRepo, mockMailer)

	email := "test@example.com"
	password := "123"
	role := "ROLE_HOMEOWNER"

	user, err := svc.RegisterUser(email, password, role)

	assert.Nil(t, user)
	assert.ErrorIs(t, err, auth.ErrWeakPassword)
}

// TestRegisterUser_InvalidEmail verifies that registration fails
// when the provided email is not a valid format.
func TestRegisterUser_InvalidEmail(t *testing.T) {
	mockRepo := new(MockUserRepo)
	mockMailer := new(MockSender)
	svc := auth.NewService(mockRepo, mockMailer)

	email := "invalid-email"
	password := "securepass"
	role := "ROLE_HOMEOWNER"

	user, err := svc.RegisterUser(email, password, role)

	assert.Nil(t, user)
	assert.ErrorIs(t, err, auth.ErrInvalidEmail)
}
