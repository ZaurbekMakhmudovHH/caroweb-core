package admin

import (
	domainuser "carowebapp/core/internal/domain/user"
	"carowebapp/core/internal/infrastructure/email"
	"errors"
	"go.uber.org/zap"
)

const (
	errMsgInvalidUserStatus    = "user cannot be approved in current status"
	errMsgUserNotFound         = "user not found"
	logMsgApprovalEmailFailed  = "failed to send approval email"
	logMsgRejectionEmailFailed = "failed to send rejection email"
	logMsgSaveRejectionFailed  = "failed to save rejection reasons"
)

var (
	ErrInvalidUserStatus = errors.New(errMsgInvalidUserStatus)
	ErrUserNotFound      = errors.New(errMsgUserNotFound)
)

// Service handles user moderation operations such as approval and rejection.
type Service struct {
	repo   Repository
	logger *zap.Logger
	Sender email.Sender
}

// NewService creates a new instance of the Service.
func NewService(repo Repository, logger *zap.Logger, sender email.Sender) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
		Sender: sender,
	}
}

// ApproveUser approves a pending user and sends them a confirmation email.
func (s *Service) ApproveUser(userID string) error {
	u, err := s.getPendingUserOrError(userID)
	if err != nil {
		return err
	}

	if err := s.repo.SetUserApproved(userID); err != nil {
		return err
	}

	if err := s.Sender.SendApprovalNotification(u.Email); err != nil {
		s.logger.Warn(logMsgApprovalEmailFailed,
			zap.String("user_id", u.ID),
			zap.String("email", u.Email),
			zap.Error(err),
		)
	}

	return nil
}

// RejectUser marks a user as rejected, stores the rejection reasons, and sends a notification.
func (s *Service) RejectUser(userID string, rejectionErrors map[string]string) error {
	u, err := s.getPendingUserOrError(userID)
	if err != nil {
		return err
	}

	if err := s.repo.InsertUserRejection(userID, rejectionErrors); err != nil {
		s.logger.Error(logMsgSaveRejectionFailed,
			zap.String("user_id", userID),
			zap.Any("errors", rejectionErrors),
			zap.Error(err),
		)
		return err
	}

	if err := s.repo.SetUserRejected(userID); err != nil {
		return err
	}

	if err := s.Sender.SendRejectionNotification(u.Email, rejectionErrors); err != nil {
		s.logger.Warn(logMsgRejectionEmailFailed,
			zap.String("user_id", u.ID),
			zap.String("email", u.Email),
			zap.Error(err),
		)
	}

	return nil
}

// ListPendingUsers retrieves users with 'pending' status and supports pagination and optional search.
func (s *Service) ListPendingUsers(search string, limit, offset int) ([]*domainuser.User, error) {
	return s.repo.ListPendingUsers(search, limit, offset)
}

// GetUserProfile returns the user's profile from the repository layer.
func (s *Service) GetUserProfile(userID string) (*domainuser.Profile, error) {
	return s.repo.GetUserProfile(userID)
}

// getPendingUserOrError fetches a user and ensures their status is 'pending'.
func (s *Service) getPendingUserOrError(userID string) (*User, error) {
	u, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	if u.Status != domainuser.StatusPending {
		return nil, ErrInvalidUserStatus
	}
	return u, nil
}
