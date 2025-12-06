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

	// GET with path parameter
	router.GET("/hello/:name", func(ctx *fox.Context) string {
		name := ctx.Param("name")
		return "Hello, " + name + "!"
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
