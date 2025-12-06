package main

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/fox-gonic/fox"
	"github.com/fox-gonic/fox/httperrors"
)

// StrongPassword validates password strength
type StrongPassword struct {
	Password string `json:"password" binding:"required"`
}

func (sp *StrongPassword) IsValid() error {
	pwd := sp.Password

	// Check minimum length
	if len(pwd) < 8 {
		return &httperrors.Error{
			HTTPCode: http.StatusBadRequest,
			Code:     "PASSWORD_TOO_SHORT",
			Err:      errors.New("password must be at least 8 characters long"),
		}
	}

	// Check for uppercase
	if !regexp.MustCompile(`[A-Z]`).MatchString(pwd) {
		return &httperrors.Error{
			HTTPCode: http.StatusBadRequest,
			Code:     "PASSWORD_NO_UPPERCASE",
			Err:      errors.New("password must contain at least one uppercase letter"),
		}
	}

	// Check for lowercase
	if !regexp.MustCompile(`[a-z]`).MatchString(pwd) {
		return &httperrors.Error{
			HTTPCode: http.StatusBadRequest,
			Code:     "PASSWORD_NO_LOWERCASE",
			Err:      errors.New("password must contain at least one lowercase letter"),
		}
	}

	// Check for digit
	if !regexp.MustCompile(`[0-9]`).MatchString(pwd) {
		return &httperrors.Error{
			HTTPCode: http.StatusBadRequest,
			Code:     "PASSWORD_NO_DIGIT",
			Err:      errors.New("password must contain at least one digit"),
		}
	}

	// Check for special character
	if !regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(pwd) {
		return &httperrors.Error{
			HTTPCode: http.StatusBadRequest,
			Code:     "PASSWORD_NO_SPECIAL",
			Err:      errors.New("password must contain at least one special character"),
		}
	}

	return nil
}

// SignupRequest with custom validation
type SignupRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (sr *SignupRequest) IsValid() error {
	// Username validation
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(sr.Username) {
		return &httperrors.Error{
			HTTPCode: http.StatusBadRequest,
			Code:     "INVALID_USERNAME",
			Err:      errors.New("username can only contain letters, numbers, underscore, and dash"),
		}
	}

	// Reserved usernames
	reserved := []string{"admin", "root", "system", "api", "www"}
	for _, r := range reserved {
		if strings.EqualFold(sr.Username, r) {
			return &httperrors.Error{
				HTTPCode: http.StatusBadRequest,
				Code:     "RESERVED_USERNAME",
				Err:      errors.New("this username is reserved"),
			}
		}
	}

	// Email domain validation
	allowedDomains := []string{"example.com", "test.com", "demo.com"}
	emailParts := strings.Split(sr.Email, "@")
	if len(emailParts) == 2 {
		domain := emailParts[1]
		valid := false
		for _, d := range allowedDomains {
			if domain == d {
				valid = true
				break
			}
		}
		if !valid {
			return &httperrors.Error{
				HTTPCode: http.StatusBadRequest,
				Code:     "INVALID_EMAIL_DOMAIN",
				Err:      errors.New("email domain not allowed. Use: " + strings.Join(allowedDomains, ", ")),
			}
		}
	}

	// Password validation
	pwdReq := &StrongPassword{Password: sr.Password}
	return pwdReq.IsValid()
}

// CreatePostRequest with content validation
type CreatePostRequest struct {
	Title   string   `json:"title" binding:"required,min=5,max=200"`
	Content string   `json:"content" binding:"required,min=10"`
	Tags    []string `json:"tags" binding:"required,min=1,max=10"`
}

func (cpr *CreatePostRequest) IsValid() error {
	// Check for profanity in title
	profanityWords := []string{"badword1", "badword2"}
	titleLower := strings.ToLower(cpr.Title)
	for _, word := range profanityWords {
		if strings.Contains(titleLower, word) {
			return &httperrors.Error{
				HTTPCode: http.StatusBadRequest,
				Code:     "PROFANITY_DETECTED",
				Err:      errors.New("title contains inappropriate content"),
			}
		}
	}

	// Validate tags
	for _, tag := range cpr.Tags {
		if len(tag) < 2 || len(tag) > 30 {
			return &httperrors.Error{
				HTTPCode: http.StatusBadRequest,
				Code:     "INVALID_TAG_LENGTH",
				Err:      errors.New("each tag must be between 2 and 30 characters"),
			}
		}

		if !regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString(tag) {
			return &httperrors.Error{
				HTTPCode: http.StatusBadRequest,
				Code:     "INVALID_TAG_FORMAT",
				Err:      errors.New("tags can only contain letters, numbers, and dashes"),
			}
		}
	}

	return nil
}

func main() {
	router := fox.New()

	// Password validation endpoint
	router.POST("/validate-password", func(_ *fox.Context, _ *StrongPassword) (string, error) {
		return "Password is strong!", nil
	})

	// Signup with comprehensive validation
	router.POST("/signup", func(_ *fox.Context, req *SignupRequest) (map[string]any, error) {
		return map[string]any{
			"message":  "Account created successfully",
			"username": req.Username,
			"email":    req.Email,
		}, nil
	})

	// Create post with content validation
	router.POST("/posts", func(_ *fox.Context, req *CreatePostRequest) (map[string]any, error) {
		return map[string]any{
			"message": "Post created successfully",
			"post": map[string]any{
				"title":   req.Title,
				"content": req.Content,
				"tags":    req.Tags,
			},
		}, nil
	})

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
