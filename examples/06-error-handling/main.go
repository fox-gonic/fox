package main

import (
	"errors"
	"net/http"

	"github.com/fox-gonic/fox"
	"github.com/fox-gonic/fox/httperrors"
)

// Custom error definitions
var (
	ErrUserNotFound = &httperrors.Error{
		HTTPCode: http.StatusNotFound,
		Code:     "USER_NOT_FOUND",
		Err:      errors.New("the requested user does not exist"),
	}

	ErrInsufficientBalance = &httperrors.Error{
		HTTPCode: http.StatusPaymentRequired,
		Code:     "INSUFFICIENT_BALANCE",
		Err:      errors.New("account balance is insufficient for this transaction"),
	}

	ErrDuplicateEmail = &httperrors.Error{
		HTTPCode: http.StatusConflict,
		Code:     "DUPLICATE_EMAIL",
		Err:      errors.New("this email address is already registered"),
	}

	ErrInvalidCredentials = &httperrors.Error{
		HTTPCode: http.StatusUnauthorized,
		Code:     "INVALID_CREDENTIALS",
		Err:      errors.New("the email or password you entered is incorrect"),
	}
)

// User represents a user model
type User struct {
	ID      int64   `json:"id"`
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	Balance float64 `json:"balance"`
}

func main() {
	router := fox.New()

	// Simple error return
	router.GET("/error/simple", func(ctx *fox.Context) (string, error) {
		return "", errors.New("something went wrong")
	})

	// HTTP error with code
	router.GET("/error/http", func(ctx *fox.Context) (string, error) {
		return "", &httperrors.Error{
			HTTPCode: http.StatusBadRequest,
			Code:     "BAD_REQUEST",
			Err:      errors.New("the request was invalid"),
		}
	})

	// Pre-defined error
	router.GET("/user/:id", func(ctx *fox.Context) (*User, error) {
		id := ctx.Param("id")

		// Simulate user not found
		if id == "999" {
			return nil, ErrUserNotFound
		}

		return &User{
			ID:      1,
			Name:    "Alice",
			Email:   "alice@example.com",
			Balance: 1000.0,
		}, nil
	})

	// Conditional error handling
	router.POST("/transfer", func(ctx *fox.Context) (map[string]interface{}, error) {
		var req struct {
			FromUserID int64   `json:"from_user_id" binding:"required"`
			ToUserID   int64   `json:"to_user_id" binding:"required"`
			Amount     float64 `json:"amount" binding:"required,gt=0"`
		}

		if err := ctx.ShouldBindJSON(&req); err != nil {
			return nil, err
		}

		// Simulate user balance check
		userBalance := 500.0
		if req.Amount > userBalance {
			return nil, ErrInsufficientBalance
		}

		// Simulate duplicate transfer check
		if req.FromUserID == req.ToUserID {
			return nil, &httperrors.Error{
				HTTPCode: http.StatusBadRequest,
				Code:     "SAME_ACCOUNT_TRANSFER",
				Err:      errors.New("cannot transfer to the same account"),
			}
		}

		return map[string]interface{}{
			"status":      "success",
			"from_user":   req.FromUserID,
			"to_user":     req.ToUserID,
			"amount":      req.Amount,
			"new_balance": userBalance - req.Amount,
		}, nil
	})

	// Login with error handling
	router.POST("/login", func(ctx *fox.Context) (map[string]interface{}, error) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := ctx.ShouldBindJSON(&req); err != nil {
			return nil, err
		}

		// Simulate credential check
		if req.Email != "alice@example.com" || req.Password != "password123" {
			return nil, ErrInvalidCredentials
		}

		return map[string]interface{}{
			"token":   "fake-jwt-token",
			"user_id": 1,
		}, nil
	})

	// Signup with duplicate check
	router.POST("/signup", func(ctx *fox.Context) (map[string]interface{}, error) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=6"`
			Name     string `json:"name" binding:"required"`
		}

		if err := ctx.ShouldBindJSON(&req); err != nil {
			return nil, err
		}

		// Simulate duplicate email check
		if req.Email == "alice@example.com" {
			return nil, ErrDuplicateEmail
		}

		return map[string]interface{}{
			"message": "Account created successfully",
			"user": map[string]interface{}{
				"email": req.Email,
				"name":  req.Name,
			},
		}, nil
	})

	// Multiple possible errors
	router.DELETE("/user/:id", func(ctx *fox.Context) (map[string]string, error) {
		id := ctx.Param("id")

		// Check if user exists
		if id == "999" {
			return nil, ErrUserNotFound
		}

		// Check if user can be deleted
		if id == "1" {
			return nil, &httperrors.Error{
				HTTPCode: http.StatusForbidden,
				Code:     "CANNOT_DELETE_ADMIN",
				Err:      errors.New("admin user cannot be deleted"),
			}
		}

		return map[string]string{
			"message": "User deleted successfully",
		}, nil
	})

	// Error with additional details
	router.GET("/detailed-error", func(ctx *fox.Context) (string, error) {
		return "", &httperrors.Error{
			HTTPCode: http.StatusBadRequest,
			Code:     "VALIDATION_FAILED",
			Err:      errors.New("request validation failed"),
			Fields: map[string]any{
				"field":  "email",
				"reason": "invalid format",
				"value":  "not-an-email",
			},
		}
	})

	// Panic recovery example
	router.GET("/panic", func(ctx *fox.Context) string {
		panic("something went wrong!")
	})

	router.Run(":8080")
}
