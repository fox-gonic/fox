package main

import (
	"errors"
	"net/http"

	"github.com/fox-gonic/fox"
	"github.com/fox-gonic/fox/httperrors"
)

// User represents a user model
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Age      int    `json:"age" binding:"gte=0,lte=150"`
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Age      int    `json:"age" binding:"gte=18,lte=150"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	ID       int64  `uri:"id" binding:"required,gt=0"`
	Username string `json:"username" binding:"omitempty,min=3,max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
}

// QueryUsersRequest represents query parameters
type QueryUsersRequest struct {
	Page     int    `form:"page" binding:"omitempty,gte=1"`
	PageSize int    `form:"page_size" binding:"omitempty,gte=1,lte=100"`
	Keyword  string `form:"keyword"`
}

func main() {
	router := fox.New()

	// POST: Create user with JSON body binding
	router.POST("/users", func(_ *fox.Context, req *CreateUserRequest) (*User, error) {
		// In real application, save to database
		user := &User{
			ID:       1,
			Username: req.Username,
			Email:    req.Email,
			Age:      req.Age,
		}
		return user, nil
	})

	// PUT: Update user with URI and JSON binding
	router.PUT("/users/:id", func(_ *fox.Context, req *UpdateUserRequest) (*User, error) {
		// In real application, update in database
		user := &User{
			ID:       req.ID,
			Username: req.Username,
			Email:    req.Email,
		}
		return user, nil
	})

	// GET: Query users with query parameters
	router.GET("/users", func(_ *fox.Context, req *QueryUsersRequest) (map[string]any, error) {
		// Set defaults
		if req.Page == 0 {
			req.Page = 1
		}
		if req.PageSize == 0 {
			req.PageSize = 10
		}

		// In real application, query from database
		return map[string]any{
			"page":      req.Page,
			"page_size": req.PageSize,
			"keyword":   req.Keyword,
			"total":     100,
			"users": []User{
				{ID: 1, Username: "alice", Email: "alice@example.com", Age: 25},
				{ID: 2, Username: "bob", Email: "bob@example.com", Age: 30},
			},
		}, nil
	})

	// GET: Get user by ID
	router.GET("/users/:id", func(ctx *fox.Context) (*User, error) {
		_ = ctx.Param("id")

		// In real application, fetch from database
		return &User{
			ID:       1,
			Username: "alice",
			Email:    "alice@example.com",
			Age:      25,
		}, nil
	})

	// Custom validation example
	router.POST("/validate", func(_ *fox.Context, req *CreateUserRequest) (string, error) {
		// Additional custom validation
		if req.Username == "admin" {
			return "", &httperrors.Error{
				HTTPCode: http.StatusBadRequest,
				Code:     "INVALID_USERNAME",
				Err:      errors.New("username 'admin' is reserved"),
			}
		}

		return "Validation passed", nil
	})

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
