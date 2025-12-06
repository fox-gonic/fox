package fox

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// ==================== Engine/Routing Benchmarks ====================

// BenchmarkEngine_SimpleRoute benchmarks a simple route with string return
func BenchmarkEngine_SimpleRoute(b *testing.B) {
	router := New()
	router.GET("/ping", func(c *Context) string {
		return "pong"
	})

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkEngine_ParamRoute benchmarks a route with path parameters
func BenchmarkEngine_ParamRoute(b *testing.B) {
	router := New()
	router.GET("/users/:id", func(c *Context) string {
		return c.Param("id")
	})

	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkEngine_MultiParam benchmarks a route with multiple path parameters
func BenchmarkEngine_MultiParam(b *testing.B) {
	router := New()
	router.GET("/users/:id/posts/:post_id", func(c *Context) map[string]string {
		return map[string]string{
			"user_id": c.Param("id"),
			"post_id": c.Param("post_id"),
		}
	})

	req := httptest.NewRequest("GET", "/users/123/posts/456", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkEngine_WildcardRoute benchmarks a wildcard route
func BenchmarkEngine_WildcardRoute(b *testing.B) {
	router := New()
	router.GET("/files/*filepath", func(c *Context) string {
		return c.Param("filepath")
	})

	req := httptest.NewRequest("GET", "/files/path/to/file.txt", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkEngine_StaticRoutes benchmarks multiple static routes
func BenchmarkEngine_StaticRoutes(b *testing.B) {
	router := New()
	router.GET("/", func(c *Context) string { return "home" })
	router.GET("/about", func(c *Context) string { return "about" })
	router.GET("/contact", func(c *Context) string { return "contact" })
	router.GET("/products", func(c *Context) string { return "products" })
	router.GET("/services", func(c *Context) string { return "services" })

	req := httptest.NewRequest("GET", "/products", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkEngine_JSONResponse benchmarks JSON response rendering
func BenchmarkEngine_JSONResponse(b *testing.B) {
	type User struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	router := New()
	router.GET("/user", func(c *Context) *User {
		return &User{
			ID:       123,
			Username: "john",
			Email:    "john@example.com",
		}
	})

	req := httptest.NewRequest("GET", "/user", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkEngine_WithMiddleware benchmarks routes with middleware
func BenchmarkEngine_WithMiddleware(b *testing.B) {
	router := New()

	// Add middleware
	router.Use(func(c *Context) {
		c.Set("middleware", true)
		c.Next()
	})

	router.GET("/ping", func(c *Context) string {
		return "pong"
	})

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkEngine_MultipleMiddlewares benchmarks routes with multiple middlewares
func BenchmarkEngine_MultipleMiddlewares(b *testing.B) {
	router := New()

	// Add multiple middlewares
	router.Use(func(c *Context) {
		c.Set("mw1", true)
		c.Next()
	})
	router.Use(func(c *Context) {
		c.Set("mw2", true)
		c.Next()
	})
	router.Use(func(c *Context) {
		c.Set("mw3", true)
		c.Next()
	})

	router.GET("/ping", func(c *Context) string {
		return "pong"
	})

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkEngine_GroupRoutes benchmarks grouped routes
func BenchmarkEngine_GroupRoutes(b *testing.B) {
	router := New()

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.GET("/users", func(c *Context) string { return "users" })
			v1.GET("/posts", func(c *Context) string { return "posts" })
		}
	}

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkEngine_POST benchmarks POST request handling
func BenchmarkEngine_POST(b *testing.B) {
	router := New()
	router.POST("/data", func(c *Context) string {
		return "received"
	})

	req := httptest.NewRequest("POST", "/data", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkEngine_AllMethods benchmarks different HTTP methods
func BenchmarkEngine_AllMethods(b *testing.B) {
	router := New()
	handler := func(c *Context) string { return "ok" }

	router.GET("/resource", handler)
	router.POST("/resource", handler)
	router.PUT("/resource", handler)
	router.DELETE("/resource", handler)
	router.PATCH("/resource", handler)

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	requests := make([]*http.Request, len(methods))
	for i, method := range methods {
		requests[i] = httptest.NewRequest(method, "/resource", nil)
	}
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, requests[i%len(requests)])
	}
}

// BenchmarkEngine_LargeRouteTable benchmarks with many routes
func BenchmarkEngine_LargeRouteTable(b *testing.B) {
	router := New()

	// Register 100 unique routes
	for i := 0; i < 100; i++ {
		path := "/route/" + string(rune('a'+i%26)) + "/" + string(rune('0'+i/26))
		router.GET(path, func(c *Context) string { return "ok" })
	}

	req := httptest.NewRequest("GET", "/route/e/0", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// ==================== Binding Benchmarks ====================

// BenchmarkBinding_URIParam benchmarks URI parameter binding
func BenchmarkBinding_URIParam(b *testing.B) {
	type Request struct {
		ID int64 `uri:"id"`
	}

	router := New()
	router.GET("/users/:id", func(c *Context, req *Request) int64 {
		return req.ID
	})

	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkBinding_QueryParam benchmarks query parameter binding
func BenchmarkBinding_QueryParam(b *testing.B) {
	type Request struct {
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
		Keyword  string `form:"keyword"`
	}

	router := New()
	router.GET("/search", func(c *Context, req *Request) map[string]interface{} {
		return map[string]interface{}{
			"page":      req.Page,
			"page_size": req.PageSize,
			"keyword":   req.Keyword,
		}
	})

	req := httptest.NewRequest("GET", "/search?page=1&page_size=10&keyword=test", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkBinding_JSONBody benchmarks JSON body binding
func BenchmarkBinding_JSONBody(b *testing.B) {
	type Request struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Age      int    `json:"age" binding:"gte=0,lte=150"`
	}

	router := New()
	router.POST("/users", func(c *Context, req *Request) *Request {
		return req
	})

	body := `{"username":"john","email":"john@example.com","age":30}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Need to reset request body for each iteration
		req.Body = io.NopCloser(strings.NewReader(body))
		router.ServeHTTP(w, req)
	}
}

// BenchmarkBinding_FormData benchmarks form data binding
func BenchmarkBinding_FormData(b *testing.B) {
	type Request struct {
		Username string `form:"username" binding:"required"`
		Email    string `form:"email" binding:"required,email"`
		Age      int    `form:"age"`
	}

	router := New()
	router.POST("/login", func(c *Context, req *Request) *Request {
		return req
	})

	formData := url.Values{}
	formData.Set("username", "john")
	formData.Set("email", "john@example.com")
	formData.Set("age", "30")
	body := formData.Encode()

	req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req.Body = io.NopCloser(strings.NewReader(body))
		router.ServeHTTP(w, req)
	}
}

// BenchmarkBinding_MixedParams benchmarks mixed parameter binding (URI + Query + JSON)
func BenchmarkBinding_MixedParams(b *testing.B) {
	type Request struct {
		ID       int64  `uri:"id"`
		Page     int    `form:"page"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	router := New()
	router.POST("/users/:id/update", func(c *Context, req *Request) *Request {
		return req
	})

	body := `{"username":"john","email":"john@example.com"}`
	req := httptest.NewRequest("POST", "/users/123/update?page=1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req.Body = io.NopCloser(strings.NewReader(body))
		router.ServeHTTP(w, req)
	}
}

// BenchmarkBinding_WithValidation benchmarks binding with validation
func BenchmarkBinding_WithValidation(b *testing.B) {
	type Request struct {
		Username string `json:"username" binding:"required,min=3,max=50"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
		Age      int    `json:"age" binding:"required,gte=18,lte=150"`
	}

	router := New()
	router.POST("/register", func(c *Context, req *Request) *Request {
		return req
	})

	body := `{"username":"john","email":"john@example.com","password":"password123","age":30}`
	req := httptest.NewRequest("POST", "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req.Body = io.NopCloser(strings.NewReader(body))
		router.ServeHTTP(w, req)
	}
}

// BenchmarkBinding_SimpleStruct benchmarks simple struct binding
func BenchmarkBinding_SimpleStruct(b *testing.B) {
	type Request struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	router := New()
	router.POST("/data", func(c *Context, req *Request) *Request {
		return req
	})

	body := `{"name":"test","value":42}`
	req := httptest.NewRequest("POST", "/data", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req.Body = io.NopCloser(strings.NewReader(body))
		router.ServeHTTP(w, req)
	}
}

// BenchmarkBinding_ComplexStruct benchmarks complex nested struct binding
func BenchmarkBinding_ComplexStruct(b *testing.B) {
	type Address struct {
		Street  string `json:"street"`
		City    string `json:"city"`
		Country string `json:"country"`
	}

	type Request struct {
		Username string   `json:"username" binding:"required"`
		Email    string   `json:"email" binding:"required,email"`
		Age      int      `json:"age" binding:"gte=0"`
		Address  Address  `json:"address"`
		Tags     []string `json:"tags"`
	}

	router := New()
	router.POST("/users", func(c *Context, req *Request) *Request {
		return req
	})

	body := `{"username":"john","email":"john@example.com","age":30,"address":{"street":"Main St","city":"NYC","country":"USA"},"tags":["developer","go"]}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req.Body = io.NopCloser(strings.NewReader(body))
		router.ServeHTTP(w, req)
	}
}

// BenchmarkBinding_NoBinding benchmarks handler without binding (baseline)
func BenchmarkBinding_NoBinding(b *testing.B) {
	router := New()
	router.POST("/data", func(c *Context) string {
		return "ok"
	})

	body := `{"username":"john","email":"john@example.com"}`
	req := httptest.NewRequest("POST", "/data", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req.Body = io.NopCloser(strings.NewReader(body))
		router.ServeHTTP(w, req)
	}
}

// BenchmarkBinding_ArrayParam benchmarks array parameter binding
func BenchmarkBinding_ArrayParam(b *testing.B) {
	type Request struct {
		IDs   []int64  `form:"ids"`
		Names []string `form:"names"`
	}

	router := New()
	router.GET("/filter", func(c *Context, req *Request) *Request {
		return req
	})

	req := httptest.NewRequest("GET", "/filter?ids=1&ids=2&ids=3&names=a&names=b", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkBinding_ReflectionOverhead benchmarks reflection overhead
func BenchmarkBinding_ReflectionOverhead(b *testing.B) {
	type Request struct {
		Field1  string  `json:"field1"`
		Field2  int     `json:"field2"`
		Field3  float64 `json:"field3"`
		Field4  bool    `json:"field4"`
		Field5  string  `json:"field5"`
		Field6  int     `json:"field6"`
		Field7  float64 `json:"field7"`
		Field8  bool    `json:"field8"`
		Field9  string  `json:"field9"`
		Field10 int     `json:"field10"`
	}

	router := New()
	router.POST("/reflect", func(c *Context, req *Request) *Request {
		return req
	})

	body := `{"field1":"a","field2":1,"field3":1.1,"field4":true,"field5":"b","field6":2,"field7":2.2,"field8":false,"field9":"c","field10":3}`
	req := httptest.NewRequest("POST", "/reflect", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req.Body = io.NopCloser(strings.NewReader(body))
		router.ServeHTTP(w, req)
	}
}
