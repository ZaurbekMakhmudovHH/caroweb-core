// Package admin provides handlers and services for administrative user management.
package admin

const (
	ErrMsgInvalidApproveBody = "invalid approve request body"
	ErrMsgInvalidRejectBody  = "invalid reject request body"

	ErrMsgApproveFailed = "failed to approve user"
	ErrMsgRejectFailed  = "failed to reject user"

	SuccessMsgApproved = "user approval completed successfully"
	SuccessMsgRejected = "user reject completed successfully"
)
