package main

import (
	"github.com/fox-gonic/fox"
)

func main() {
	// Create a new Fox engine
	router := fox.New()

	// Simple GET endpoint
	router.GET("/ping", func() string {
		return "pong"
	})

	// GET with path parameter - using ctx.Param()
	router.GET("/hello/:name", func(ctx *fox.Context) string {
		name := ctx.Param("name")
		return "Hello, " + name + "!"
	})

	// GET with path parameter - using struct binding
	type UserParams struct {
		Name string `uri:"name" binding:"required"`
	}
	router.GET("/greet/:name", func(params UserParams) string {
		return "Greetings, " + params.Name + "!"
	})

	// POST endpoint returning JSON
	router.POST("/echo", func() map[string]string {
		return map[string]string{
			"message": "Echo service is working",
		}
	})

	// Start server on port 8080
	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
