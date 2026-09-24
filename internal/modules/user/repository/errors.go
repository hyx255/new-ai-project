package repository

import "errors"

// ErrNotFound is returned when a requested entity does not exist.
var ErrNotFound = errors.New("not found")

// ErrUsernameExists is returned when a username already exists.
var ErrUsernameExists = errors.New("username already exists")