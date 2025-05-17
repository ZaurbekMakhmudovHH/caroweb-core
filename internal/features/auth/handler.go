package auth

import (
	"carowebapp/core/internal/infrastructure/response"

	"carowebapp/core/internal/pkg/contextutils"

	"errors"

	"github.com/go-playground/validator/v10"

	"github.com/gofiber/fiber/v2"

	"go.uber.org/zap"

	"time"
)

type Handler struct {
	service   *Service
	validator *validator.Validate
	logger    *zap.Logger
}

func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service:   service,
		validator: validator.New(),
		logger:    logger,
	}
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type CreateProfileRequest struct {
	Salutation  string `json:"salutation" validate:"required"`
	Title       string `json:"title"`
	FirstName   string `json:"firstName" validate:"required"`
	LastName    string `json:"lastName" validate:"required"`
	Street      string `json:"street" validate:"required"`
	HouseNumber string `json:"houseNumber" validate:"required"`
	PostalCode  string `json:"postalCode" validate:"required"`
	City        string `json:"city" validate:"required"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordPayload struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

func (h *Handler) Register(c *fiber.Ctx) error {
	req, _ := contextutils.GetValidatedBody[RegisterRequest](c)

	user, err := h.service.RegisterUser(req.Email, req.Password, req.Role)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidEmail), errors.Is(err, ErrWeakPassword), errors.Is(err, ErrInvalidRole):
			return response.JSONErrorInfoLog(c, h.logger, fiber.StatusBadRequest, err.Error(),
				zap.String("email", req.Email),
				zap.String("reason", err.Error()),
			)

		case errors.Is(err, ErrEmailExists):
			return response.JSONErrorWithLog(c, h.logger, fiber.StatusConflict, err.Error(),
				zap.String("email", req.Email),
			)

		default:
			return response.JSONErrorWithLog(c, h.logger, fiber.StatusInternalServerError, err.Error(),
				zap.Error(err),
				zap.String("email", req.Email),
			)
		}
	}

	logFields := []zap.Field{
		zap.String("user_id", user.ID),
		zap.String("email", user.Email),
	}
	if traceID, ok := contextutils.GetTraceID(c); ok {
		logFields = append(logFields, zap.String("trace_id", traceID))
	}

	h.logger.Info("User registered", logFields...)

	return response.JSONSuccess(c, fiber.StatusCreated, fiber.Map{
		"id":    user.ID,
		"email": user.Email,
	})
}

func (h *Handler) Login(c *fiber.Ctx) error {
	req, _ := contextutils.GetValidatedBody[LoginRequest](c)

	user, accessToken, refreshToken, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		return response.JSONErrorInfoLog(c, h.logger, fiber.StatusUnauthorized, response.ErrMsgLoginFailed,
			zap.String("email", req.Email),
			zap.Error(err),
		)
	}

	h.logger.Info("User logged in",
		zap.String("user_id", user.ID),
		zap.String("email", user.Email),
	)

	return response.JSONSuccess(c, fiber.StatusOK, fiber.Map{
		"access_token":    accessToken,
		"refresh_token":   refreshToken,
		"email_confirmed": user.EmailConfirmed,
		"role":            user.Role,
		"status":          user.Status,
	})
}

func (h *Handler) ConfirmEmail(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return response.JSONErrorInfoLog(c, h.logger, fiber.StatusBadRequest, response.ErrMsgMissingToken,
			zap.String("path", c.Path()),
			zap.String("method", c.Method()),
		)
	}

	user, err := h.service.GetByConfirmationToken(token)
	if err != nil || user == nil {
		return response.JSONErrorInfoLog(c, h.logger, fiber.StatusBadRequest, response.ErrMsgInvalidOrExpiredToken,
			zap.String("token", token),
			zap.Error(err),
		)
	}

	if err := h.service.ConfirmEmail(user.ID); err != nil {
		return response.JSONErrorWithLog(c, h.logger, fiber.StatusInternalServerError, response.ErrMsgEmailConfirmationFail,
			zap.String("user_id", user.ID),
			zap.Error(err),
		)
	}

	h.logger.Info("Email confirmed successfully",
		zap.String("user_id", user.ID),
		zap.String("email", user.Email),
	)

	return response.JSONSuccess(c, fiber.StatusOK, fiber.Map{
		"message": "email confirmed successfully",
	})
}

func (h *Handler) ResendConfirmation(c *fiber.Ctx) error {
	userID, ok := contextutils.GetUserID(c)
	if !ok {
		return response.JSONErrorInfoLog(c, h.logger, fiber.StatusUnauthorized, response.ErrMsgUnauthorized)
	}

	err := h.service.ResendConfirmationByUserID(userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrAlreadyConfirmed):
			return response.JSONErrorInfoLog(c, h.logger, fiber.StatusBadRequest, response.ErrMsgAlreadyConfirmed,
				zap.String("user_id", userID),
			)

		default:
			return response.JSONErrorInfoLog(c, h.logger, fiber.StatusTooManyRequests, err.Error(),
				zap.String("user_id", userID),
				zap.Error(err),
			)
		}
	}

	h.logger.Info("Confirmation email re-sent",
		zap.String("user_id", userID),
	)

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) CreateProfile(c *fiber.Ctx) error {
	req, _ := contextutils.GetValidatedBody[CreateProfileRequest](c)
	userID, ok := contextutils.GetUserID(c)
	if !ok {
		return response.JSONErrorInfoLog(c, h.logger, fiber.StatusUnauthorized, response.ErrMsgUnauthorized)
	}

	var titlePtr *string
	if req.Title != "" {
		titlePtr = &req.Title
	}

	profile := &UserProfile{
		UserID:      userID,
		Salutation:  req.Salutation,
		Title:       titlePtr,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Street:      req.Street,
		HouseNumber: req.HouseNumber,
		PostalCode:  req.PostalCode,
		City:        req.City,
		IsVerified:  false,
		UpdatedAt:   time.Now(),
	}

	if err := h.service.AddUserProfile(profile); err != nil {
		return response.JSONErrorWithLog(c, h.logger, fiber.StatusInternalServerError, response.ErrMsgProfileCreationFail,
			zap.String("user_id", userID),
			zap.Error(err),
		)
	}

	h.logger.Info("User profile created",
		zap.String("user_id", userID),
	)

	return c.SendStatus(fiber.StatusCreated)
}

func (h *Handler) RequestResetPassword(c *fiber.Ctx) error {
	req, _ := contextutils.GetValidatedBody[ResetPasswordRequest](c)

	if err := h.service.RequestPasswordReset(req.Email); err != nil {
		h.logger.Info("Password reset requested", zap.String("email", req.Email), zap.Error(err))
		return c.SendStatus(fiber.StatusNoContent)
	}

	h.logger.Info("Password reset email sent", zap.String("email", req.Email))
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) CheckResetPasswordToken(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return response.JSONErrorInfoLog(c, h.logger, fiber.StatusBadRequest, "missing token")
	}

	valid, err := h.service.IsResetTokenValid(token)
	if err != nil || !valid {
		return response.JSONErrorInfoLog(c, h.logger, fiber.StatusBadRequest, "invalid or expired token", zap.Error(err))
	}

	return response.JSONSuccess(c, fiber.StatusOK, fiber.Map{
		"message": "Token is valid",
	})
}

func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	req, _ := contextutils.GetValidatedBody[ResetPasswordPayload](c)

	if err := h.service.ResetPassword(req.Token, req.NewPassword); err != nil {
		return response.JSONErrorWithLog(c, h.logger, fiber.StatusBadRequest, err.Error(), zap.Error(err))
	}

	h.logger.Info("Password reset successful", zap.String("token", req.Token))
	return response.JSONSuccess(c, fiber.StatusOK, fiber.Map{
		"message": "password has been reset",
	})
}
