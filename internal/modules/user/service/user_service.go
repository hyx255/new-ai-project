package service

import (
	"context"
	"errors"
	"fmt"

	"broadcast-platform/internal/modules/user/model"
	"broadcast-platform/internal/modules/user/repository"
	"broadcast-platform/internal/platform/logging"

	"go.uber.org/zap"
)

type UserService struct {
	userRepo repository.UserRepository
	logger   *logging.Logger
}

func NewUserService(userRepo repository.UserRepository, logger *logging.Logger) *UserService {
	return &UserService{userRepo: userRepo, logger: logger}
}

type CreateUserRequest struct {
	Username        string
	InitialPassword string
	Role            model.UserRole
}

func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*model.SafeUser, error) {
	if err := RequireAdmin(ctx); err != nil {
		return nil, err
	}
	if req.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if len(req.Username) < 2 || len(req.Username) > 64 {
		return nil, fmt.Errorf("username must be between 2 and 64 characters")
	}
	if !req.Role.IsValid() {
		return nil, fmt.Errorf("invalid role: must be ADMIN or USER")
	}
	if err := NewPasswordPolicy().Validate(req.InitialPassword); err != nil {
		return nil, err
	}
	hash, err := HashPassword(req.InitialPassword)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	user := &model.User{
		ID:                 model.GenerateUserID(),
		Username:           req.Username,
		PasswordHash:       hash,
		Role:               req.Role,
		Status:             model.StatusActive,
		MustChangePassword: true,
		TokenVersion:       1,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrUsernameExists) {
			return nil, fmt.Errorf("username already exists")
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	s.logger.Info("user created", zap.String("user_id", user.ID), zap.String("username", user.Username))
	safe := user.ToSafe()
	return &safe, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*model.SafeUser, error) {
	if err := RequireAdmin(ctx); err != nil {
		return nil, err
	}
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	safe := user.ToSafe()
	if user.IsDeleted() {
		safe.Username = safe.Username + " (deleted)"
	}
	return &safe, nil
}

type ListUsersRequest struct {
	Page           int
	PageSize       int
	IncludeDeleted bool
}

type ListUsersResponse struct {
	Users []model.SafeUser `json:"users"`
	Total int              `json:"total"`
	Page  int              `json:"page"`
	Size  int              `json:"page_size"`
}

func (s *UserService) ListUsers(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error) {
	if err := RequireAdmin(ctx); err != nil {
		return nil, err
	}
	users, total, err := s.userRepo.List(ctx, req.Page, req.PageSize, req.IncludeDeleted)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	safeUsers := make([]model.SafeUser, 0, len(users))
	for _, u := range users {
		safe := u.ToSafe()
		if u.IsDeleted() {
			safe.Username = safe.Username + " (deleted)"
		}
		safeUsers = append(safeUsers, safe)
	}
	return &ListUsersResponse{
		Users: safeUsers,
		Total: total,
		Page:  req.Page,
		Size:  req.PageSize,
	}, nil
}

type UpdateUserRequest struct {
	Username string
	Role     model.UserRole
}

func (s *UserService) UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (*model.SafeUser, error) {
	if err := RequireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := PreventSelfAction(ctx, id); err != nil {
		return nil, err
	}
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user.IsDeleted() {
		return nil, fmt.Errorf("user not found")
	}
	if req.Username != "" {
		if len(req.Username) < 2 || len(req.Username) > 64 {
			return nil, fmt.Errorf("username must be between 2 and 64 characters")
		}
		user.Username = req.Username
	}
	if req.Role.IsValid() {
		user.Role = req.Role
	}
	if err := s.userRepo.Update(ctx, user); err != nil {
		if errors.Is(err, repository.ErrUsernameExists) {
			return nil, fmt.Errorf("username already exists")
		}
		return nil, fmt.Errorf("update user: %w", err)
	}
	s.logger.Info("user updated", zap.String("user_id", user.ID))
	safe := user.ToSafe()
	return &safe, nil
}

type UpdateUserStatusRequest struct {
	Status model.UserStatus
}

func (s *UserService) UpdateUserStatus(ctx context.Context, id string, req UpdateUserStatusRequest) error {
	if err := RequireAdmin(ctx); err != nil {
		return err
	}
	if err := PreventSelfAction(ctx, id); err != nil {
		return err
	}
	if !req.Status.IsValid() {
		return fmt.Errorf("invalid status: must be ACTIVE or DISABLED")
	}
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("get user: %w", err)
	}
	if user.IsDeleted() {
		return fmt.Errorf("user not found")
	}
	if user.Status == req.Status {
		return fmt.Errorf("user is already %s", req.Status)
	}
	newTokenVersion := user.TokenVersion + 1
	if err := s.userRepo.UpdateStatus(ctx, id, req.Status, newTokenVersion); err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	s.logger.Info("user status changed", zap.String("user_id", id), zap.String("new_status", string(req.Status)))
	return nil
}

type ResetPasswordRequest struct {
	NewPassword string
}

func (s *UserService) ResetPassword(ctx context.Context, id string, req ResetPasswordRequest) error {
	if err := RequireAdmin(ctx); err != nil {
		return err
	}
	if err := PreventSelfAction(ctx, id); err != nil {
		return err
	}
	if err := NewPasswordPolicy().Validate(req.NewPassword); err != nil {
		return err
	}
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("get user: %w", err)
	}
	if user.IsDeleted() {
		return fmt.Errorf("user not found")
	}
	hash, err := HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	newTokenVersion := user.TokenVersion + 1
	if err := s.userRepo.UpdatePassword(ctx, id, hash, true, newTokenVersion); err != nil {
		return fmt.Errorf("reset password: %w", err)
	}
	s.logger.Info("password reset by admin", zap.String("user_id", id))
	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	if err := RequireAdmin(ctx); err != nil {
		return err
	}
	if err := PreventSelfAction(ctx, id); err != nil {
		return err
	}
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("get user: %w", err)
	}
	if user.IsDeleted() {
		return fmt.Errorf("user not found")
	}
	if err := s.userRepo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	s.logger.Info("user deleted", zap.String("user_id", id))
	return nil
}