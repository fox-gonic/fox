# Fox Web 框架

[English](README.md) | 简体中文

[![Go Tests](https://github.com/fox-gonic/fox/actions/workflows/go.yml/badge.svg)](https://github.com/fox-gonic/fox/actions/workflows/go.yml)
[![Security Scanning](https://github.com/fox-gonic/fox/actions/workflows/security.yml/badge.svg)](https://github.com/fox-gonic/fox/actions/workflows/security.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/fox-gonic/fox)](https://goreportcard.com/report/github.com/fox-gonic/fox)
[![GoDoc](https://pkg.go.dev/badge/github.com/fox-gonic/fox?status.svg)](https://pkg.go.dev/github.com/fox-gonic/fox)
[![codecov](https://codecov.io/gh/fox-gonic/fox/branch/main/graph/badge.svg)](https://codecov.io/gh/fox-gonic/fox)

Fox 是 [Gin](https://github.com/gin-gonic/gin) Web 框架的强大扩展，提供自动参数绑定、灵活的响应渲染和增强功能，同时保持与 Gin 的完全兼容。

## 特性

- 🚀 **自动绑定和渲染**: 自动绑定请求参数并渲染响应
- 🔧 **Handler 灵活性**: 支持多种 Handler 签名，自动类型检测
- 🌐 **多域名路由**: 基于域名的流量路由，支持精确匹配和正则表达式
- ✅ **自定义验证**: 实现 `IsValider` 接口以支持复杂验证逻辑
- 📊 **结构化日志**: 内置日志系统，支持 TraceID、结构化字段和文件轮转
- ⚡ **高性能**: 在 Gin 快速路由基础上增加最小开销
- 🔒 **安全优先**: 内置安全扫描和最佳实践
- 📦 **100% Gin 兼容**: 无缝使用任何 Gin 中间件或功能

## 目录

- [安装](#安装)
- [快速开始](#快速开始)
- [架构](#架构)
- [性能](#性能)
- [示例](#示例)
- [最佳实践](#最佳实践)
- [故障排查](#故障排查)
- [安全](#安全)
- [贡献](#贡献)
- [许可证](#许可证)

## ⚠️ **注意**

Fox 目前处于 beta 阶段，正在积极开发中。虽然它提供了令人兴奋的新功能，但请注意它可能不适合生产环境使用。如果您选择使用，请做好应对潜在 bug 和破坏性变更的准备。始终查看官方文档和发布说明以获取更新，并谨慎使用。祝编码愉快！

## 安装

Fox 需要 **Go 版本 `1.24` 或更高**。如果需要安装或升级 Go，请访问 [Go 官方下载页面](https://go.dev/dl/)。首先为您的项目创建一个新目录并进入该目录。然后，在终端中执行以下命令，使用 Go modules 初始化您的项目：

```bash
go mod init github.com/your/repo
```

要了解更多关于 Go modules 的信息，可以查看 [使用 Go Modules](https://go.dev/blog/using-go-modules) 博客文章。

设置好项目后，可以使用 `go get` 命令安装 Fox：

```bash
go get -u github.com/fox-gonic/fox
```

此命令会获取 Fox 包并将其添加到项目依赖中，让您可以开始使用 Fox 构建 Web 应用程序。

## 快速开始

### 运行 Fox Engine

首先需要导入 fox 包以使用 fox engine，最简单的示例如下 `example.go`：

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
  router.Run() // 监听并服务于 0.0.0.0:8080 (Windows 为 "localhost:8080")
}
```

使用 Go 命令运行示例：

```shell
# 运行 example.go 并在浏览器访问 0.0.0.0:8080/ping
$ go run example.go
```

### 自动绑定请求数据并渲染

```go
package main

import (
  "github.com/fox-gonic/fox"
)

type DescribeArticleArgs struct {
  ID int64 `uri:"id"`
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
    article := &Article{
      Title:     args.Title,
      Content:   args.Content,
      CreatedAt: time.Now(),
      UpdatedAt: time.Now(),
    }
    // 保存文章到数据库
    return article, nil
  })

  router.Run()
}
```

#### 支持自定义 IsValider 进行绑定验证

```go
package main

import (
  "github.com/fox-gonic/fox"
)

var ErrPasswordTooShort = &httperrors.Error{
	HTTPCode: http.StatusBadRequest,
	Err:      errors.New("password too short"),
	Code:     "PASSWORD_TOO_SHORT",
}

type CreateUserArgs struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (args *CreateUserArgs) IsValid() error {
	if args.Username == "" && args.Email == "" {
		return httperrors.ErrInvalidArguments
	}
	if len(args.Password) < 6 {
		return ErrPasswordTooShort
	}
	return nil
}

func main() {
  router := fox.New()

  router.POST("/users/signup", func(c *fox.Context, args *CreateUserArgs) (*User, error) {
    user := &User{
      Username: args.Username,
      Email:    args.Email,
    }
    // 对密码进行哈希并保存用户到数据库
    return user, nil
  })

  router.Run()
}
```

```shell
$ curl -X POST http://localhost:8080/users/signup \
    -H 'content-type: application/json' \
    -d '{"username": "George", "email": "george@vandaley.com"}'
{"code":"PASSWORD_TOO_SHORT"}
```

## 架构

Fox 扩展了 Gin 的路由引擎，增加了自动参数绑定和响应渲染功能：

```
┌─────────────────────────────────────────────────────────────┐
│                         HTTP 请求                            │
└────────────────────────────┬────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                      Gin 路由/引擎                           │
│  ┌────────────────┐  ┌──────────────┐  ┌─────────────────┐ │
│  │   中间件 1     │─▶│   中间件 2   │─▶│   中间件 N     │ │
│  └────────────────┘  └──────────────┘  └─────────────────┘ │
└────────────────────────────┬────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                     Fox Handler 包装器                       │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  1. 反射 Handler 签名                                │  │
│  │     • 检测参数类型 (Context, Request 等)             │  │
│  │     • 检测返回类型 (data, error, status)             │  │
│  └──────────────────────────────────────────────────────┘  │
│                             │                                │
│                             ▼                                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  2. 自动参数绑定                                     │  │
│  │     • URI 参数 (路径变量)                            │  │
│  │     • Query 参数                                     │  │
│  │     • JSON/Form 请求体                               │  │
│  │     • 自定义验证 (IsValider)                         │  │
│  └──────────────────────────────────────────────────────┘  │
│                             │                                │
│                             ▼                                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  3. 执行 Handler 函数                                │  │
│  │     • 使用绑定的参数调用                             │  │
│  │     • 处理 panic 和错误                              │  │
│  └──────────────────────────────────────────────────────┘  │
│                             │                                │
│                             ▼                                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  4. 自动响应渲染                                     │  │
│  │     • 检测响应类型                                   │  │
│  │     • 序列化为 JSON                                  │  │
│  │     • 设置适当的 HTTP 状态码                         │  │
│  │     • 特殊处理 httperrors.Error                      │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                        HTTP 响应                             │
└─────────────────────────────────────────────────────────────┘
```

### 核心组件

- **fox.Engine**: 包装 `gin.Engine` 并增强 Handler 注册
- **fox.Context**: 扩展 `gin.Context` 并添加额外方法 (RequestBody, TraceID)
- **call.go**: 基于反射的核心 Handler 调用逻辑
- **render.go**: 自动响应序列化和渲染
- **validator.go**: 集成 go-playground/validator 和自定义 IsValider
- **DomainEngine**: 多域名路由，支持精确匹配和正则表达式模式

## 性能

Fox 在 Gin 的性能基础上增加了最小开销，同时显著提升了开发效率：

### 基准测试对比

```
BenchmarkGin_SimpleRoute         10000000    118 ns/op    0 B/op    0 allocs/op
BenchmarkFox_SimpleRoute         10000000    125 ns/op    0 B/op    0 allocs/op
BenchmarkFox_AutoBinding          5000000    312 ns/op  128 B/op    3 allocs/op
BenchmarkFox_AutoRendering        8000000    187 ns/op   64 B/op    2 allocs/op
```

### 性能特征

| 功能 | 开销 | 说明 |
|------|------|------|
| 简单路由 (字符串返回) | ~6% | 每次路由注册的一次性反射成本 |
| 自动绑定 (结构体参数) | ~165% | 包括 JSON 解析和验证 |
| 自动渲染 (结构体返回) | ~58% | 包括 JSON 序列化 |
| 复杂 Handler | ~10-20% | 在请求处理过程中均摊 |

**关键洞察**: 开销主要来自 JSON 解析/序列化，而非 Fox 的反射逻辑。对于大多数实际应用，相比数据库查询和业务逻辑，这些开销可以忽略不计。

### 何时使用 Fox vs Gin

**使用 Fox 当**:
- 构建具有多个端点的 REST API
- 需要自动参数验证
- 希望更简洁、更易维护的 Handler 签名
- 处理 JSON 请求/响应体

**直接使用 Gin 当**:
- 每一微秒都很重要（高频交易等）
- 需要对请求/响应处理的最大控制
- 构建静态文件服务器或代理

## 示例

在 [`examples/`](../examples/) 目录中提供了全面的示例：

| 示例 | 描述 |
|------|------|
| [01-basic](../examples/01-basic) | 基础路由、路径参数、JSON 响应 |
| [02-binding](../examples/02-binding) | 参数绑定 (JSON/URI/Query) 和验证 |
| [03-middleware](../examples/03-middleware) | 自定义中间件、身份验证、限流 |
| [04-domain-routing](../examples/04-domain-routing) | 多域名和多租户路由 |
| [05-custom-validator](../examples/05-custom-validator) | 使用 IsValider 接口的复杂验证 |
| [06-error-handling](../examples/06-error-handling) | HTTP 错误、自定义错误码 |
| [07-logger-config](../examples/07-logger-config) | 日志配置、文件轮转、JSON 日志 |

每个示例都包含带有使用说明和 curl 命令的 README。

## 最佳实践

### 1. 错误处理

**使用 httperrors.Error 处理 API 错误：**

```go
import "github.com/fox-gonic/fox/httperrors"

var ErrUserNotFound = &httperrors.Error{
    HTTPCode: http.StatusNotFound,
    Code:     "USER_NOT_FOUND",
    Err:      errors.New("user not found"),
}

router.GET("/users/:id", func(ctx *fox.Context) (*User, error) {
    user, err := findUser(ctx.Param("id"))
    if err != nil {
        return nil, ErrUserNotFound
    }
    return user, nil
})
```

### 2. 请求验证

**结合结构体标签和 IsValider：**

```go
type CreateUserRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
    Age      int    `json:"age" binding:"gte=18,lte=150"`
}

func (r *CreateUserRequest) IsValid() error {
    if strings.Contains(r.Email, "disposable.com") {
        return &httperrors.Error{
            HTTPCode: http.StatusBadRequest,
            Code:     "INVALID_EMAIL_DOMAIN",
            Err:      errors.New("不允许使用一次性邮箱地址"),
        }
    }
    return nil
}
```

### 3. 结构化日志

**使用带字段的 logger 以获得更好的可观测性：**

```go
import "github.com/fox-gonic/fox/logger"

router.POST("/orders", func(ctx *fox.Context, req *CreateOrderRequest) (*Order, error) {
    log := logger.NewWithContext(ctx.Context)

    log.WithFields(map[string]interface{}{
        "user_id": req.UserID,
        "amount":  req.Amount,
    }).Info("Creating order")

    order, err := createOrder(req)
    if err != nil {
        log.WithError(err).Error("Order creation failed")
        return nil, err
    }

    return order, nil
})
```

### 4. Handler 签名

**根据使用场景选择正确的签名：**

```go
// 简单: 不需要绑定
router.GET("/health", func(ctx *fox.Context) string {
    return "OK"
})

// 带绑定: 自动参数提取
router.GET("/users/:id", func(ctx *fox.Context, req *GetUserRequest) (*User, error) {
    return findUser(req.ID)
})

// 完全控制: 访问上下文并返回自定义状态
router.POST("/complex", func(ctx *fox.Context, req *Request) (interface{}, int, error) {
    result, err := process(req)
    if err != nil {
        return nil, http.StatusInternalServerError, err
    }
    return result, http.StatusCreated, nil
})
```

### 5. 生产环境配置

**为生产环境配置日志：**

```go
import "github.com/fox-gonic/fox/logger"

logger.SetConfig(&logger.Config{
    LogLevel:              logger.InfoLevel,
    ConsoleLoggingEnabled: true,
    FileLoggingEnabled:    true,
    Filename:              "/var/log/myapp/app.log",
    MaxSize:               100,  // MB
    MaxBackups:            30,
    MaxAge:                90,   // 天数
    EncodeLogsAsJSON:      true,
})

router := fox.New()
router.Use(fox.Logger(fox.LoggerConfig{
    SkipPaths: []string{"/health", "/metrics"},
}))
```

### 6. 多域名路由

**按域名组织路由：**

```go
de := fox.NewDomainEngine()

// API 子域名
de.Domain("api.example.com", func(apiRouter *fox.Engine) {
    apiRouter.GET("/v1/users", listUsers)
    apiRouter.POST("/v1/users", createUser)
})

// Admin 子域名
de.Domain("admin.example.com", func(adminRouter *fox.Engine) {
    adminRouter.Use(AuthMiddleware())
    adminRouter.GET("/dashboard", showDashboard)
})

// 租户子域名通配符
de.DomainRegexp(`^(?P<tenant>[a-z0-9-]+)\.example\.com$`, func(tenantRouter *fox.Engine) {
    tenantRouter.GET("/", func(ctx *fox.Context) string {
        tenant := ctx.Param("tenant")
        return "欢迎, " + tenant
    })
})

http.ListenAndServe(":8080", de)
```

## 故障排查

### 常见问题

#### 1. 绑定验证失败

**问题**: 请求验证失败，错误消息不清晰。

**解决方案**: 检查结构体标签并正确使用 `binding` 标签：

```go
// 错误
type Request struct {
    Email string `json:"email" validate:"email"`  // 错误的标签
}

// 正确
type Request struct {
    Email string `json:"email" binding:"required,email"`
}
```

#### 2. Handler 未找到 / 404 错误

**问题**: 即使已注册路由，仍然返回 404。

**解决方案**:
- 确保路径参数匹配: `/users/:id` vs `/users/:user_id`
- 检查 HTTP 方法: `GET` vs `POST`
- 如果使用 DomainEngine，验证域名路由配置
- 启用调试模式查看已注册的路由:

```go
fox.SetMode(fox.DebugMode)
```

#### 3. JSON 解析错误

**问题**: `invalid character` 或 `cannot unmarshal` 错误。

**解决方案**:
- 验证 Content-Type header 是 `application/json`
- 检查 JSON 结构是否匹配结构体标签
- 使用正确的字段类型 (string vs int)

```bash
# 正确
curl -H "Content-Type: application/json" -d '{"name":"Alice"}' http://localhost:8080/users

# 缺少 header (可能失败)
curl -d '{"name":"Alice"}' http://localhost:8080/users
```

#### 4. 自定义验证器未调用

**问题**: `IsValid()` 方法未被调用。

**解决方案**: 确保使用指针接收器和正确的接口：

```go
// 正确
func (r *CreateUserRequest) IsValid() error {
    return nil
}

// 错误 (值接收器不起作用)
func (r CreateUserRequest) IsValid() error {
    return nil
}
```

#### 5. 域名路由中的正则表达式 Panic

**问题**: 注册无效正则表达式的域名时应用程序 panic。

**解决方案**: 在注册前验证正则表达式模式：

```go
pattern := `^(?P<tenant>[a-z0-9-]+)\.example\.com$`
if _, err := regexp.Compile(pattern); err != nil {
    log.Fatal("Invalid regex:", err)
}
de.DomainRegexp(pattern, handler)
```

#### 6. 内存占用过高

**问题**: 内存使用随时间增长。

**可能原因**:
- 日志文件句柄未关闭 (检查 MaxBackups/MaxAge)
- 大响应体未被垃圾回收
- 中间件内存泄漏

**解决方案**:
```go
// 正确配置日志轮转
logger.SetConfig(&logger.Config{
    MaxBackups: 10,   // 仅保留 10 个旧文件
    MaxAge:     30,   // 删除超过 30 天的文件
})

// 为长时间运行的请求使用上下文超时
ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
defer cancel()
```

### 调试模式

启用调试模式查看详细信息：

```go
fox.SetMode(fox.DebugMode)  // 开发环境
fox.SetMode(fox.ReleaseMode)  // 生产环境
```

在调试模式下，Fox 会打印：
- 已注册的路由及其 Handler
- 请求绑定详情
- 中间件执行顺序

### 获取帮助

1. 查看 [examples/](../examples/) 目录
2. 阅读 [CONTRIBUTING.md](../CONTRIBUTING.md) 了解指南
3. 搜索现有的 [GitHub Issues](https://github.com/fox-gonic/fox/issues)
4. 提交新 issue 时包含:
   - Fox 和 Go 版本
   - 最小可复现示例
   - 预期行为与实际行为对比

## 安全

Fox 非常重视安全性。我们实施了多层安全扫描：

### 自动化安全扫描

- **govulncheck**: 扫描 Go 依赖中的已知漏洞
- **CodeQL**: 静态应用安全测试 (SAST) 进行代码分析
- **Dependency Review**: 审查 Pull Request 中的依赖变更
- **每周扫描**: 每周一自动运行安全扫描

### 本地运行安全扫描

```bash
# 安装 govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# 运行漏洞扫描
govulncheck ./...
```

### 安全文档

- [SECURITY.md](../SECURITY.md) - 安全策略和漏洞报告
- [SECURITY_SCAN.md](.github/SECURITY_SCAN.md) - 详细的安全扫描文档

### 报告安全问题

如果您发现安全漏洞，请参阅 [SECURITY.md](../SECURITY.md) 了解我们的负责任披露流程。**不要**为安全漏洞提交公开的 GitHub issue。

## 贡献

我们欢迎贡献！请查看 [CONTRIBUTING.md](../CONTRIBUTING.md) 了解如何为 Fox 做出贡献的详细信息。

## 许可证

Fox 使用 MIT 许可证发布。详见 [LICENSE](../LICENSE)。
