package service

import (
	"context"
	"errors"
	"fmt"

	"broadcast-platform/internal/modules/user/model"
	"broadcast-platform/internal/modules/user/repository"
	"broadcast-platform/internal/platform/auth"
	"broadcast-platform/internal/platform/logging"

	"go.uber.org/zap"
)

// AuthService handles authentication logic.
type AuthService struct {
	userRepo      repository.UserRepository
	tokenProvider *auth.TokenProvider
	logger        *logging.Logger
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	userRepo repository.UserRepository,
	tokenProvider *auth.TokenProvider,
	logger *logging.Logger,
) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		tokenProvider: tokenProvider,
		logger:        logger,
	}
}

// LoginRequest represents login input.
type LoginRequest struct {
	Username string
	Password string
}

// LoginResponse represents login output.
type LoginResponse struct {
	AccessToken string
	User        model.SafeUser
}

// Login authenticates a user and returns a JWT.
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Find user by username
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// Don't distinguish between user not found and wrong password
			return nil, fmt.Errorf("invalid username or password")
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	// Check if deleted
	if user.IsDeleted() {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Check if disabled
	if user.Status != model.StatusActive {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Verify password
	if !ComparePassword(user.PasswordHash, req.Password) {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Generate JWT
	token, err := s.tokenProvider.Generate(
		user.ID,
		user.Username,
		string(user.Role),
		user.MustChangePassword,
		user.TokenVersion,
	)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	s.logger.Info("user logged in",
		zap.String("user_id", user.ID),
		zap.String("username", user.Username),
	)

	return &LoginResponse{
		AccessToken: token,
		User:        user.ToSafe(),
	}, nil
}

// GetCurrentUser retrieves the current authenticated user from context.
func (s *AuthService) GetCurrentUser(ctx context.Context) (*model.User, error) {
	currentUser := GetCurrentUser(ctx)
	if currentUser == nil {
		return nil, fmt.Errorf("no authenticated user in context")
	}

	user, err := s.userRepo.GetByID(ctx, currentUser.ID)
	if err != nil {
		return nil, fmt.Errorf("get current user: %w", err)
	}

	return user, nil
}

// ChangePasswordRequest represents change password input.
type ChangePasswordRequest struct {
	CurrentPassword string
	NewPassword     string
}

// ChangePassword changes the current user password.
func (s *AuthService) ChangePassword(ctx context.Context, req ChangePasswordRequest) error {
	currentUser := GetCurrentUser(ctx)
	if currentUser == nil {
		return fmt.Errorf("no authenticated user in context")
	}

	// Load user from database
	user, err := s.userRepo.GetByID(ctx, currentUser.ID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	// Verify current password
	if !ComparePassword(user.PasswordHash, req.CurrentPassword) {
		return fmt.Errorf("current password is incorrect")
	}

	// Check new password != old password
	if req.NewPassword == req.CurrentPassword {
		return fmt.Errorf("new password must be different from current password")
	}

	// Validate new password
	if err := NewPasswordPolicy().Validate(req.NewPassword); err != nil {
		return err
	}

	// Hash new password
	newHash, err := HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// Update password, clear must_change_password, increment token_version
	newTokenVersion := user.TokenVersion + 1
	if err := s.userRepo.UpdatePassword(ctx, user.ID, newHash, false, newTokenVersion); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	s.logger.Info("password changed",
		zap.String("user_id", user.ID),
		zap.String("username", user.Username),
	)

	return nil
}