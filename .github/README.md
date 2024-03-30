# Fox Web Framework

The fox is an extension of the [gin](https://github.com/gin-gonic/gin) framework

## ⚠️ **Attention**

Fox is currently in beta and under active development. While it offers exciting new features, please note that it may not be stable for production use. If you choose to use, be prepared for potential bugs and breaking changes. Always check the official documentation and release notes for updates and proceed with caution. Happy coding!


## Installation

Fox requires **Go version `1.21` or higher** to run. If you need to install or upgrade Go, visit the [official Go download page](https://go.dev/dl/). To start setting up your project. Create a new directory for your project and navigate into it. Then, initialize your project with Go modules by executing the following command in your terminal:

```bash
go mod init github.com/your/repo
```

To learn more about Go modules and how they work, you can check out the [Using Go Modules](https://go.dev/blog/using-go-modules) blog post.

After setting up your project, you can install fox with the `go get` command:

```bash
go get -u github.com/fox-gonic/fox
```

This command fetches the Fox package and adds it to your project's dependencies, allowing you to start building your web applications with Fox.


## Quickstart

### Running fox Engine

First you need to import fox package for using fox engine, one simplest example likes the follow `example.go`:

```go
package main

import (
  "github.com/fox-gonic/fox"
)

func main() {
  router := fox.New()
  router.GET("/ping", func(c *fox.Context) string {
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
  "github.com/fox-gonic/fox"
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
  router := fox.New()

  router.GET("/articles/:id", func(c *fox.Context, args *DescribeArticleArgs) int64 {
    return args.ID
  })

  router.POST("/articles", func(c *fox.Context, args *CreateArticleArgs) (*Article, error) {
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
