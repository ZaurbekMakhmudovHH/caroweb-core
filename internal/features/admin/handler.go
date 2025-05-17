package admin

import (
	domainuser "carowebapp/core/internal/domain/user"

	"carowebapp/core/internal/infrastructure/response"

	"carowebapp/core/internal/pkg/contextutils"

	"github.com/gofiber/fiber/v2"

	"go.uber.org/zap"
)

// Handler provides HTTP handlers for user moderation endpoints.
// Assumes admin access is already enforced via middleware.
type Handler struct {
	Service      *Service
	Logger       *zap.Logger
	UserProvider domainuser.Provider
}

// ApproveUserRequest represents the payload for approving a user.
type ApproveUserRequest struct {
	UserID string `json:"user_id" validate:"required,uuid4"`
}

// RejectUserRequest represents the payload for rejecting a user with specific field errors.
type RejectUserRequest struct {
	UserID string            `json:"user_id" validate:"required,uuid4"`
	Errors map[string]string `json:"errors" validate:"required"`
}

// ApproveUser allows an admin to approve a pending user by their user ID.
func (h *Handler) ApproveUser(c *fiber.Ctx) error {
	req, ok := contextutils.GetValidatedBody[ApproveUserRequest](c)
	if !ok {
		h.Logger.Debug(ErrMsgInvalidApproveBody, zap.String("handler", "ApproveUser"))
		return response.JSONError(c, fiber.StatusBadRequest, fiber.ErrBadRequest)
	}

	if err := h.Service.ApproveUser(req.UserID); err != nil {
		return response.JSONErrorWithLog(c, h.Logger, fiber.StatusInternalServerError, ErrMsgApproveFailed,
			zap.String("target_user_id", req.UserID),
			zap.Error(err),
		)
	}

	h.Logger.Info(SuccessMsgApproved,
		zap.String("target_user_id", req.UserID),
	)

	return response.JSONSuccess(c, fiber.StatusOK, fiber.Map{
		"user_id": req.UserID,
		"status":  domainuser.StatusApproved,
	})
}

// RejectUser allows an admin to reject a user by providing field-level validation errors.
func (h *Handler) RejectUser(c *fiber.Ctx) error {
	req, ok := contextutils.GetValidatedBody[RejectUserRequest](c)
	if !ok {
		h.Logger.Debug(ErrMsgInvalidRejectBody, zap.String("handler", "RejectUser"))
		return response.JSONError(c, fiber.StatusBadRequest, fiber.ErrBadRequest)
	}

	if err := h.Service.RejectUser(req.UserID, req.Errors); err != nil {
		return response.JSONErrorWithLog(c, h.Logger, fiber.StatusInternalServerError, ErrMsgRejectFailed,
			zap.String("target_user_id", req.UserID),
			zap.Error(err),
			zap.Any("rejection_errors", req.Errors),
		)
	}

	h.Logger.Info(SuccessMsgRejected,
		zap.String("target_user_id", req.UserID),
		zap.Any("rejection_errors", req.Errors),
	)

	return response.JSONSuccess(c, fiber.StatusOK, fiber.Map{
		"user_id": req.UserID,
		"status":  domainuser.StatusRejected,
	})
}

// ListPendingUsers returns a paginated list of users with 'pending' status.
// Supports optional search by first or last name via ?search=... query param.
func (h *Handler) ListPendingUsers(c *fiber.Ctx) error {
	search := c.Query("search", "")
	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	limit := 25
	offset := (page - 1) * limit

	users, err := h.Service.ListPendingUsers(search, limit, offset)
	if err != nil {
		return response.JSONErrorWithLog(c, h.Logger, fiber.StatusInternalServerError, "failed to list pending users",
			zap.String("search", search),
			zap.Int("page", page),
			zap.Error(err),
		)
	}

	h.Logger.Info("Pending users fetched",
		zap.Int("count", len(users)),
		zap.String("search", search),
		zap.Int("page", page),
	)

	return response.JSONSuccess(c, fiber.StatusOK, fiber.Map{
		"page":  page,
		"limit": limit,
		"total": len(users),
		"users": users,
	})
}

// GetUserProfile returns the profile of a specific user by their ID.
func (h *Handler) GetUserProfile(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		h.Logger.Debug("missing user ID in path", zap.String("handler", "GetUserProfile"))
		return response.JSONError(c, fiber.StatusBadRequest, fiber.ErrBadRequest)
	}

	profile, err := h.Service.GetUserProfile(userID)
	if err != nil {
		return response.JSONErrorWithLog(c, h.Logger, fiber.StatusInternalServerError, "failed to get user profile",
			zap.String("user_id", userID),
			zap.Error(err),
		)
	}
	if profile == nil {
		return response.JSONErrorInfoLog(c, h.Logger, fiber.StatusNotFound, "user profile not found",
			zap.String("user_id", userID),
		)
	}

	h.Logger.Info("User profile fetched",
		zap.String("user_id", userID),
	)

	return response.JSONSuccess(c, fiber.StatusOK, profile)
}
