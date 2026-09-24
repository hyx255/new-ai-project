package errors

import (
	"net/http"
	"testing"
)

func TestNewSystemError(t *testing.T) {
	err := NewSystemError("system error")

	if err.Code != CodeSystemError {
		t.Errorf("expected code %d, got %d", CodeSystemError, err.Code)
	}

	if err.Message != "system error" {
		t.Errorf("expected message 'system error', got '%s'", err.Message)
	}

	if err.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, err.HTTPStatus)
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("validation failed")

	if err.Code != CodeValidationError {
		t.Errorf("expected code %d, got %d", CodeValidationError, err.Code)
	}

	if err.HTTPStatus != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, err.HTTPStatus)
	}
}

func TestNewBusinessError(t *testing.T) {
	err := NewBusinessError(2001, "business rule violated")

	if err.Code != 2001 {
		t.Errorf("expected code 2001, got %d", err.Code)
	}

	if err.Message != "business rule violated" {
		t.Errorf("expected message 'business rule violated', got '%s'", err.Message)
	}
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("resource not found")

	if err.Code != CodeNotFound {
		t.Errorf("expected code %d, got %d", CodeNotFound, err.Code)
	}

	if err.HTTPStatus != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, err.HTTPStatus)
	}
}

func TestAppErrorError(t *testing.T) {
	err := NewSystemError("test error")

	expected := "[1000] test error"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}

	errWithDetails := err.WithDetails("additional details")
	expectedWithDetails := "[1000] test error: additional details"
	if errWithDetails.Error() != expectedWithDetails {
		t.Errorf("expected '%s', got '%s'", expectedWithDetails, errWithDetails.Error())
	}
}

func TestIsAppError(t *testing.T) {
	appErr := NewSystemError("test")
	if !IsAppError(appErr) {
		t.Error("expected IsAppError to return true for AppError")
	}

	regularErr := error(nil)
	if IsAppError(regularErr) {
		t.Error("expected IsAppError to return false for nil error")
	}
}

func TestAsAppError(t *testing.T) {
	appErr := NewSystemError("test")
	converted, ok := AsAppError(appErr)

	if !ok {
		t.Error("expected AsAppError to return true for AppError")
	}

	if converted != appErr {
		t.Error("expected AsAppError to return the same error")
	}
}
