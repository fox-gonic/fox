package main

import (
	"github.com/fox-gonic/fox"
)

func main() {
	// Create a new Fox engine
	router := fox.New()

	// Simple GET endpoint
	router.GET("/ping", func(ctx *fox.Context) string {
		return "pong"
	})

	// GET with path parameter
	router.GET("/hello/:name", func(ctx *fox.Context) string {
		name := ctx.Param("name")
		return "Hello, " + name + "!"
	})

	// POST endpoint returning JSON
	router.POST("/echo", func(ctx *fox.Context) map[string]string {
		return map[string]string{
			"message": "Echo service is working",
		}
	})

	// Start server on port 8080
	router.Run(":8080")
}
