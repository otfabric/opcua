package server

import (
	"context"

	"github.com/otfabric/opcua/ua"
)

// AccessController defines the interface for controlling access to server operations.
//
// Each method is called before the corresponding service executes. Return
// [ua.StatusOK] to allow the operation, or an appropriate status code
// (e.g., [ua.StatusBadUserAccessDenied]) to deny it.
//
// The session parameter provides the authenticated session context, which
// can be used to implement role-based or user-based access control.
//
// Provide a custom implementation via [WithAccessController]. The default
// is [DefaultAccessController] which allows all operations.
type AccessController interface {
	CheckRead(ctx context.Context, session *session, nodeID *ua.NodeID) ua.StatusCode
	CheckWrite(ctx context.Context, session *session, nodeID *ua.NodeID) ua.StatusCode
	CheckBrowse(ctx context.Context, session *session, nodeID *ua.NodeID) ua.StatusCode
	CheckCall(ctx context.Context, session *session, methodID *ua.NodeID) ua.StatusCode
}

// DefaultAccessController is a permissive access controller that allows all
// operations. It is used when no custom [AccessController] is configured.
type DefaultAccessController struct{}

func (DefaultAccessController) CheckRead(_ context.Context, _ *session, _ *ua.NodeID) ua.StatusCode {
	return ua.StatusOK
}

func (DefaultAccessController) CheckWrite(_ context.Context, _ *session, _ *ua.NodeID) ua.StatusCode {
	return ua.StatusOK
}

func (DefaultAccessController) CheckBrowse(_ context.Context, _ *session, _ *ua.NodeID) ua.StatusCode {
	return ua.StatusOK
}

func (DefaultAccessController) CheckCall(_ context.Context, _ *session, _ *ua.NodeID) ua.StatusCode {
	return ua.StatusOK
}

// WithAccessController sets a custom access controller on the server.
func WithAccessController(ac AccessController) Option {
	return func(s *serverConfig) {
		s.accessController = ac
	}
}
