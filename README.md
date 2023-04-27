# fox

## Engine

The fox Engine is an extension of the [gin](https://github.com/gin-gonic/gin) framework

### Running fox Engine

First you need to import fox package for using fox engine, one simplest example likes the follow `example.go`:

```go
package main

import (
  "github.com/fox-gonic/fox/engine"
)

func main() {
  router := engine.New()
  router.GET("/ping", func(c *engine.Context) string {
	  return "pong"
  })
  router.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
```

And use the Go command to run the demo:

```
# run example.go and visit 0.0.0.0:8080/ping on browser
$ go run example.go
```

### Automatically bind request data and render

```go
package main

import (
  "github.com/fox-gonic/fox/engine"
)

type DescribeArticleArgs struct {
	ID int64 `uri:"title"`
}

type CreateArticleArgs struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Article struct {
	Title     string    `json:"title"`
	Content   string    `json:"content"`
  CreatedAt time.Time `json:"created_at"`
  UpdatedAt time.Time `json:"updated_at"`
}

func main() {
  router := engine.New()

  router.GET("/articles/:id", func(c *engine.Context, args *DescribeArticleArgs) int64 {
	  return args.ID
  })

  router.POST("/articles", func(c *engine.Context, args *CreateArticleArgs) (*Article, error) {
    var article = &Article{
      Title:   args.Title,
      Content: args.Content,
    }
    // TODO: do something ...
	  return article, nil
  })

  router.Run()
}
```