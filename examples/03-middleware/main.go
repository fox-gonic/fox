package main

import (
	"strconv"
	"time"

	"github.com/fox-gonic/fox"
	"github.com/fox-gonic/fox/logger"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware simulates authentication
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "unauthorized",
				"code":  "MISSING_TOKEN",
			})
			return
		}

		if token != "Bearer valid-token" {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "unauthorized",
				"code":  "INVALID_TOKEN",
			})
			return
		}

		// Set user info in context
		c.Set("user_id", 123)
		c.Set("username", "alice")

		c.Next()
	}
}

// RateLimitMiddleware simulates rate limiting
func RateLimitMiddleware() gin.HandlerFunc {
	// In production, use a real rate limiter
	lastRequest := make(map[string]time.Time)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if last, exists := lastRequest[clientIP]; exists {
			if time.Since(last) < time.Second {
				c.AbortWithStatusJSON(429, gin.H{
					"error": "rate limit exceeded",
					"code":  "RATE_LIMIT",
				})
				return
			}
		}

		lastRequest[clientIP] = time.Now()
		c.Next()
	}
}

// RequestIDMiddleware adds request ID to context
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = strconv.FormatInt(time.Now().UnixNano(), 10)
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

func main() {
	router := fox.New()

	// Global middlewares
	router.Use(fox.Logger())            // Request logging
	router.Use(fox.NewXResponseTimer()) // Response time tracking
	router.Use(RequestIDMiddleware())   // Request ID
	router.Use(gin.Recovery())          // Panic recovery

	// Public routes (no authentication)
	router.GET("/", func() string {
		return "Welcome to Fox API"
	})

	router.GET("/health", func() map[string]string {
		return map[string]string{
			"status": "healthy",
		}
	})

	// Protected routes (with authentication)
	protected := router.Group("/api")
	protected.Use(AuthMiddleware())
	{
		protected.GET("/profile", func(ctx *fox.Context) map[string]any {
			userID, _ := ctx.Get("user_id")
			username, _ := ctx.Get("username")

			return map[string]any{
				"user_id":  userID,
				"username": username,
			}
		})

		protected.GET("/data", func(ctx *fox.Context) map[string]any {
			requestID, _ := ctx.Get("request_id")

			return map[string]any{
				"request_id": requestID,
				"data":       []string{"item1", "item2", "item3"},
			}
		})
	}

	// Rate limited routes
	limited := router.Group("/limited")
	limited.Use(RateLimitMiddleware())
	{
		limited.GET("/resource", func() string {
			return "Rate limited resource"
		})
	}

	// Custom middleware for specific route
	router.GET("/special", func(c *gin.Context) {
		c.Set("special", true)
		c.Next()
	}, func(ctx *fox.Context) map[string]any {
		special, _ := ctx.Get("special")
		return map[string]any{
			"special": special,
			"message": "This route has custom middleware",
		}
	})

	// Logger configuration example
	router.GET("/with-logger", fox.Logger(fox.LoggerConfig{
		SkipPaths: []string{"/health"},
	}), func(ctx *fox.Context) string {
		// Get logger from context
		log := logger.NewWithContext(ctx.Context)
		log.Info("Processing request with custom logger config")
		return "Logged"
	})

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
