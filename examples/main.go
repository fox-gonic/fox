package main

import (
	"fmt"

	"github.com/miclle/fox"
)

// Index home page
func Index(c *fox.Context) {
	fmt.Fprint(c.Writer, "Welcome!\n")
}

// Hello route
func Hello(c *fox.Context) {
	fmt.Fprintf(c.Writer, "hello, %s!\n", c.Params.ByName("name"))
}

// Posts route
func Posts(c *fox.Context) {
	fmt.Fprint(c.Writer, "Posts page!\n")
}

// Post route
func Post(c *fox.Context) {
	fmt.Fprintf(c.Writer, "Post detail page, id = %s!\n", c.Params.ByName("id"))
}

func main() {
	router := fox.New()
	router.GET("/", Index)
	router.GET("/hello/:name", Hello)

	group := router.Group("/api")
	group.GET("/posts", Posts)
	group.GET("/posts/:id", Post)

	router.Run(":8080")
}
