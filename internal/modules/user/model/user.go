package model

import "time"

// UserRole represents user roles.
type UserRole string

const (
	RoleAdmin UserRole = "ADMIN"
	RoleUser  UserRole = "USER"
)

// IsValid checks if the role is valid.
func (r UserRole) IsValid() bool {
	return r == RoleAdmin || r == RoleUser
}

// UserStatus represents user account status.
type UserStatus string

const (
	StatusActive   UserStatus = "ACTIVE"
	StatusDisabled UserStatus = "DISABLED"
)

// IsValid checks if the status is valid.
func (s UserStatus) IsValid() bool {
	return s == StatusActive || s == StatusDisabled
}

// User represents a user entity.
type User struct {
	ID                 string     `json:"id"`
	Username           string     `json:"username"`
	PasswordHash       string     `json:"-"`
	Role               UserRole   `json:"role"`
	Status             UserStatus `json:"status"`
	MustChangePassword bool       `json:"must_change_password"`
	TokenVersion       int        `json:"token_version"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// IsDeleted returns true if the user has been soft-deleted.
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

// IsActive returns true if the user is active and not deleted.
func (u *User) IsActive() bool {
	return !u.IsDeleted() && u.Status == StatusActive
}

// SafeUser returns a copy of the user without sensitive fields for API responses.
type SafeUser struct {
	ID                 string     `json:"id"`
	Username           string     `json:"username"`
	Role               UserRole   `json:"role"`
	Status             UserStatus `json:"status"`
	MustChangePassword bool       `json:"must_change_password"`
	TokenVersion       int        `json:"token_version"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// ToSafe converts User to SafeUser for API responses.
func (u *User) ToSafe() SafeUser {
	return SafeUser{
		ID:                 u.ID,
		Username:           u.Username,
		Role:               u.Role,
		Status:             u.Status,
		MustChangePassword: u.MustChangePassword,
		TokenVersion:       u.TokenVersion,
		DeletedAt:          u.DeletedAt,
		CreatedAt:          u.CreatedAt,
		UpdatedAt:          u.UpdatedAt,
	}
}