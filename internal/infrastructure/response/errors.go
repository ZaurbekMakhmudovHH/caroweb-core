package response

const (
	ErrMsgUnauthorized          = "unauthorized"
	ErrMsgMissingToken          = "missing token"
	ErrMsgInvalidOrExpiredToken = "invalid or expired token"
	ErrMsgEmailConfirmationFail = "could not confirm email"
	ErrMsgProfileCreationFail   = "failed to create user profile"
	ErrMsgLoginFailed           = "login failed"
	ErrMsgAlreadyConfirmed      = "email already confirmed"
)
