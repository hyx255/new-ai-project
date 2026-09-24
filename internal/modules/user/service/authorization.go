package service

import (
	"context"
	"fmt"

	"broadcast-platform/internal/modules/user/model"
)

// contextKey is a private type for context keys in this package.
type contextKey string

const currentUserKey contextKey = "current_user"

// CurrentUser represents the authenticated user extracted from JWT.
type CurrentUser struct {
	ID                 string
	Username           string
	Role               model.UserRole
	MustChangePassword bool
	TokenVersion       int
}

// WithCurrentUser stores the current user in the context.
func WithCurrentUser(ctx context.Context, user *CurrentUser) context.Context {
	return context.WithValue(ctx, currentUserKey, user)
}

// GetCurrentUser retrieves the current user from the context.
func GetCurrentUser(ctx context.Context) *CurrentUser {
	user, ok := ctx.Value(currentUserKey).(*CurrentUser)
	if !ok {
		return nil
	}
	return user
}

// RequireAdmin checks that the current user has ADMIN role.
// Returns nil if authorized, or an error if not.
func RequireAdmin(ctx context.Context) error {
	user := GetCurrentUser(ctx)
	if user == nil {
		return fmt.Errorf("unauthorized: no authenticated user")
	}
	if user.Role != model.RoleAdmin {
		return fmt.Errorf("forbidden: admin role required")
	}
	return nil
}

// PreventSelfAction ensures the current user cannot perform an action on themselves.
// Used to prevent admin self-disable and self-delete.
func PreventSelfAction(ctx context.Context, targetUserID string) error {
	user := GetCurrentUser(ctx)
	if user == nil {
		return fmt.Errorf("unauthorized: no authenticated user")
	}
	if user.ID == targetUserID {
		return fmt.Errorf("forbidden: cannot perform this action on yourself")
	}
	return nil
}
