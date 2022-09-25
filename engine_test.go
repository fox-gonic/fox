package fox

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fox-gonic/fox/render"
	"github.com/stretchr/testify/assert"
)

// PerformRequest router test
func PerformRequest(r http.Handler, method, path string, header http.Header, body ...io.Reader) *httptest.ResponseRecorder {
	var data io.Reader
	if len(body) > 0 {
		data = body[0]
	}
	req := httptest.NewRequest(method, path, data)
	req.Header = header
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestEngineStore(t *testing.T) {
	router := Default()
	router.Store("name", "value")

	router.GET("/", func(c *Context) interface{} {
		v, ok := c.Engine().Load("name")
		assert.True(t, ok)
		assert.Equal(t, "value", v)
		return v
	})

	value, exists := router.Load("name")
	assert.NotNil(t, value)
	assert.True(t, exists)

	value = router.MustLoad("name")
	assert.NotNil(t, value)

	value, exists = router.Load("foo")
	assert.Nil(t, value)
	assert.False(t, exists)

	assert.Panics(t, func() { router.MustLoad("foo") })

	w := PerformRequest(router, http.MethodGet, "/", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "value", w.Body.String())
}

func TestEngineAddRoute(t *testing.T) {
	router := Default()
	router.addRoute("GET", "/", HandlersChain{func() {}})

	assert.Len(t, router.trees, 1)
	assert.NotNil(t, router.trees.get("GET"))
	assert.Nil(t, router.trees.get("POST"))

	router.addRoute("POST", "/", HandlersChain{func() {}})

	assert.Len(t, router.trees, 2)
	assert.NotNil(t, router.trees.get("GET"))
	assert.NotNil(t, router.trees.get("POST"))

	router.addRoute("POST", "/post", HandlersChain{func() {}})
	assert.Len(t, router.trees, 2)
}

func TestEngineRegisterRoute(t *testing.T) {
	assert := assert.New(t)
	router := Default()

	var index = func(c *Context) string { return "home page" }
	var ping = func() string { return "pong" }

	type HelloHandlerArgs struct {
		Name string `pos:"path:name"`
	}
	var hello = func(c *Context, args *HelloHandlerArgs) string {
		return fmt.Sprintf("hello %s", args.Name)
	}

	var groupMiddleware = func(c *Context) {
		c.Set("groupMiddleware", "groupMiddleware")
	}

	var resources = func(c *Context) []string {
		groupMiddleware := c.MustGet("groupMiddleware").(string)
		assert.Equal(groupMiddleware, "groupMiddleware")
		return []string{"resource1", "resource2", "resource3"}
	}

	var resourceCreate = func(c *Context) []string {
		groupMiddleware := c.MustGet("groupMiddleware").(string)
		assert.Equal(groupMiddleware, "groupMiddleware")
		return []string{"resource1", "resource2", "resource3"}
	}

	type ResourceHandlerArgs struct {
		ID int `pos:"path:id"`
	}
	type Resource struct {
		ID int `json:"id"`
	}
	var resource = func(c *Context, args *ResourceHandlerArgs) Resource {
		return Resource{ID: args.ID}
	}

	router.GET("/", index)
	router.GET("/ping", ping)
	router.GET("/hello/:name", hello)

	group := router.Group("/group", groupMiddleware)
	group.GET("/resources", resources)
	group.POST("/resources", resourceCreate)
	group.GET("/resources/:id", resource)

	w := PerformRequest(router, http.MethodGet, "/", nil)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal(`home page`, w.Body.String())

	w = PerformRequest(router, http.MethodGet, "/ping", nil)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal(`pong`, w.Body.String())

	w = PerformRequest(router, http.MethodGet, "/hello/fox", nil)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("hello fox", w.Body.String())

	w = PerformRequest(router, http.MethodGet, "/group/resources", nil)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal(`["resource1","resource2","resource3"]`, w.Body.String())

	w = PerformRequest(router, http.MethodGet, "/group/resources/1", nil)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal(`{"id":1}`, w.Body.String())
}

func TestRouter(t *testing.T) {
	assert := assert.New(t)

	router := Default()
	router.GET("/GET", func(c *Context) (string, int) { return "get", 200 })
	router.HEAD("/GET", func(c *Context) (string, int) { return "head", 200 })
	router.OPTIONS("/GET", func(c *Context) (string, int) { return "options", 200 })
	router.POST("/POST", func(c *Context) (string, int) { return "post", 200 })
	router.PUT("/PUT", func(c *Context) (string, int) { return "put", 200 })
	router.PATCH("/PATCH", func(c *Context) (string, int) { return "patch", 200 })
	router.DELETE("/DELETE", func(c *Context) (string, int) { return "delete", 200 })

	router.Handle(http.MethodGet, "/user/:name", func(c *Context) string {
		want := &Params{Param{"name", "gopher"}}
		assert.Equal(c.Params, want)
		return c.Params.ByName("name")
	})

	w := PerformRequest(router, http.MethodGet, "/GET", nil)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("get", w.Body.String())

	w = PerformRequest(router, http.MethodHead, "/GET", nil)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("head", w.Body.String())

	w = PerformRequest(router, http.MethodOptions, "/GET", nil)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("options", w.Body.String())

	w = PerformRequest(router, http.MethodPost, "/POST", nil)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("post", w.Body.String())

	w = PerformRequest(router, http.MethodPut, "/PUT", nil)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("put", w.Body.String())

	w = PerformRequest(router, http.MethodPatch, "/PATCH", nil)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("patch", w.Body.String())

	w = PerformRequest(router, http.MethodDelete, "/DELETE", nil)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("delete", w.Body.String())

	w = PerformRequest(router, http.MethodGet, "/user/gopher", nil)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("gopher", w.Body.String())
}

func TestEngineRESTful(t *testing.T) {
	assert := assert.New(t)
	router := Default()

	type Product struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Desc string `json:"desc"`
	}

	type ListProductArgs struct {
		Page     int `pos:"query:page"`
		PageSize int `pos:"query:page_size"`
	}
	var index = func(c *Context, args *ListProductArgs) ([]Product, error) {
		products := make([]Product, 10)
		return products, nil
	}

	type CreateProductArgs struct {
		Name string `json:"name"`
		Desc string `json:"desc"`
	}
	var create = func(c *Context, args *CreateProductArgs) (interface{}, error) {
		product := &Product{
			ID:   1,
			Name: args.Name,
			Desc: args.Desc,
		}

		return render.JSON{
			Status: 201,
			Data:   product,
		}, nil
	}

	type GetProductArgs struct {
		ID int `pos:"path:id"`
	}
	var show = func(c *Context, args *GetProductArgs) (*Product, error) {
		if args.ID == 0 {
			return nil, &Error{
				Status: 404,
				Err:    errors.New("not found"),
			}
		}
		product := &Product{
			ID:   args.ID,
			Name: "Product Name",
			Desc: "Product Desc",
		}
		return product, nil
	}

	type UpdateProductArgs struct {
		ID   int    `pos:"path:id"`
		Name string `json:"name"`
		Desc string `json:"desc"`
	}
	var update = func(c *Context, args *UpdateProductArgs) (*Product, error) {
		product := &Product{
			ID:   args.ID,
			Name: args.Name,
			Desc: args.Desc,
		}
		return product, nil
	}

	type DestroyProductArgs struct {
		ID int `pos:"path:id"`
	}
	var destroy = func(c *Context, args *DestroyProductArgs) (*Product, error) {
		if args.ID == 0 {
			return nil, &Error{
				Status: 404,
				Err:    errors.New("not found"),
			}
		}
		return nil, nil
	}

	router.GET("/products", index)
	router.POST("/products", create)
	router.GET("/products/:id", show)
	router.PATCH("/products/:id", update)
	router.DELETE("/products/:id", destroy)

	w := PerformRequest(router, http.MethodGet, "/products", nil)
	assert.Equal(http.StatusOK, w.Code)
	var response []Product
	json.Unmarshal(w.Body.Bytes(), &response) // nolint: errcheck
	assert.Equal(10, len(response))

	body := `{
		"name": "Product Name",
		"desc": "Product Desc"
	}`
	w = PerformRequest(router, http.MethodPost, "/products", nil, strings.NewReader(body))
	assert.Equal(http.StatusCreated, w.Code)
	assert.Equal(`{"id":1,"name":"Product Name","desc":"Product Desc"}`, w.Body.String())

	for i := 0; i < 5; i++ {
		w = PerformRequest(router, http.MethodGet, fmt.Sprintf("/products/%d", i), nil)
		if i == 0 {
			assert.Equal(http.StatusNotFound, w.Code)
		} else {
			assert.Equal(http.StatusOK, w.Code)
			assert.Equal(fmt.Sprintf(`{"id":%d,"name":"Product Name","desc":"Product Desc"}`, i), w.Body.String())
		}
	}

	body = `{
		"name": "Product Name[updated]",
		"desc": "Product Desc[updated]"
	}`
	w = PerformRequest(router, http.MethodPatch, "/products/1", nil, strings.NewReader(body))
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal(`{"id":1,"name":"Product Name[updated]","desc":"Product Desc[updated]"}`, w.Body.String())

	for i := 0; i < 5; i++ {
		w = PerformRequest(router, http.MethodDelete, fmt.Sprintf("/products/%d", i), nil)
		if i == 0 {
			assert.Equal(http.StatusNotFound, w.Code)
		} else {
			assert.Equal(http.StatusOK, w.Code)
		}
	}
}

// mixed static and wildcard
func TestRouterStatic(t *testing.T) {
	router := Default()

	router.GET("/articles", func(c *Context) (string, error) {
		return "articles", nil
	})

	type Args struct {
		ID int `pos:"path:id"`
	}
	router.GET("/articles/:id", func(c *Context, args *Args) (string, error) {
		return fmt.Sprintf("articles/%d", args.ID), nil
	})

	router.GET("/articles/popular", func(c *Context) (string, error) {
		return "articles/popular", nil
	})

	w := PerformRequest(router, http.MethodGet, "/articles", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "articles", w.Body.String())

	w = PerformRequest(router, http.MethodGet, "/articles/123", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "articles/123", w.Body.String())

	w = PerformRequest(router, http.MethodGet, "/articles/popular", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "articles/popular", w.Body.String())
}

func TestRouterInvalidInput(t *testing.T) {
	router := Default()
	router.basePath = ""
	handle := func(*Context) {}
	assert.Panics(t, func() { router.Handle("", "/", handle) }, "registering empty method did not panic")
	assert.Panics(t, func() { router.GET("", handle) }, "registering empty path did not panic")
	assert.Panics(t, func() { router.GET("noSlashRoot", handle) }, "registering path not beginning with '/' did not panic")
	assert.Panics(t, func() { router.GET("/", nil) }, "registering nil handler did not panic")
}

func TestRouteRedirectTrailingSlash(t *testing.T) {
	router := Default()
	router.RedirectFixedPath = false
	router.RedirectTrailingSlash = true
	router.GET("/path", func(c *Context) {})
	router.GET("/path2/", func(c *Context) {})
	router.POST("/path3", func(c *Context) {})
	router.PUT("/path4/", func(c *Context) {})

	w := PerformRequest(router, http.MethodGet, "/path/", nil)
	assert.Equal(t, "/path", w.Header().Get("Location"))
	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path2", nil)
	assert.Equal(t, "/path2/", w.Header().Get("Location"))
	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	w = PerformRequest(router, http.MethodPost, "/path3/", nil)
	assert.Equal(t, "/path3", w.Header().Get("Location"))
	assert.Equal(t, http.StatusPermanentRedirect, w.Code)

	w = PerformRequest(router, http.MethodPut, "/path4", nil)
	assert.Equal(t, "/path4/", w.Header().Get("Location"))
	assert.Equal(t, http.StatusPermanentRedirect, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path2/", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	w = PerformRequest(router, http.MethodPost, "/path3", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	w = PerformRequest(router, http.MethodPut, "/path4/", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	header := http.Header{}
	header.Add("X-Forwarded-Prefix", "/api")
	w = PerformRequest(router, http.MethodGet, "/path2", header)
	assert.Equal(t, "/api/path2/", w.Header().Get("Location"))
	assert.Equal(t, 301, w.Code)

	header = http.Header{}
	header.Add("X-Forwarded-Prefix", "/api/")
	w = PerformRequest(router, http.MethodGet, "/path2/", header)
	assert.Equal(t, 200, w.Code)

	router.RedirectTrailingSlash = false

	w = PerformRequest(router, http.MethodGet, "/path/", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path2", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)

	w = PerformRequest(router, http.MethodPost, "/path3/", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)

	w = PerformRequest(router, http.MethodPut, "/path4", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRouteNotAllowedEnabled(t *testing.T) {
	router := Default()
	router.DefaultContentType = MIMEPlain
	router.POST("/path", func(c *Context) {})

	w := PerformRequest(router, http.MethodGet, "/path", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	router.NoMethod(func(c *Context) (interface{}, error) {
		return "responseText", &Error{Status: http.StatusTeapot}
	})

	w = PerformRequest(router, http.MethodGet, "/path", nil)
	assert.Equal(t, http.StatusText(http.StatusTeapot), w.Body.String())
	assert.Equal(t, http.StatusTeapot, w.Code)
}

func TestRouteNotAllowedEnabled2(t *testing.T) {
	router := Default()

	router.GET("/path2", func(c *Context) {})

	w := PerformRequest(router, http.MethodPost, "/path2", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestRouteNotAllowedDisabled(t *testing.T) {
	router := Default()
	router.HandleMethodNotAllowed = false
	router.POST("/path", func(c *Context) {})

	w := PerformRequest(router, http.MethodGet, "/path", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)

	router.NoMethod(func(c *Context) (interface{}, int) {
		return "responseText", http.StatusTeapot
	})

	w = PerformRequest(router, http.MethodGet, "/path", nil)
	assert.Equal(t, "404 page not found", w.Body.String())
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRouterNotFound(t *testing.T) {
	assert := assert.New(t)
	handlerFunc := func(*Context) {}

	router := Default()
	router.RedirectFixedPath = true
	router.GET("/path", handlerFunc)
	router.GET("/dir/", handlerFunc)
	router.GET("/", handlerFunc)

	testRoutes := []struct {
		route    string
		code     int
		location string
	}{
		{"/path/", http.StatusMovedPermanently, "/path"},   // TSR -/
		{"/dir", http.StatusMovedPermanently, "/dir/"},     // TSR +/
		{"/PATH", http.StatusMovedPermanently, "/path"},    // Fixed Case
		{"/DIR/", http.StatusMovedPermanently, "/dir/"},    // Fixed Case
		{"/PATH/", http.StatusMovedPermanently, "/path"},   // Fixed Case -/
		{"/DIR", http.StatusMovedPermanently, "/dir/"},     // Fixed Case +/
		{"/../path", http.StatusMovedPermanently, "/path"}, // CleanPath
		{"/nope", http.StatusNotFound, ""},                 // NotFound
	}
	for _, tr := range testRoutes {
		w := PerformRequest(router, http.MethodGet, tr.route, nil)
		assert.Equal(tr.code, w.Code)
		if w.Code != http.StatusNotFound {
			assert.Equal(tr.location, fmt.Sprint(w.Header().Get("Location")))
		}
	}

	// Test custom not found handler
	var notFound bool
	router.NotFound(func(c *Context) (interface{}, int) {
		notFound = true
		return nil, 404
	})

	w := PerformRequest(router, http.MethodGet, "/nope", nil)
	assert.Equal(http.StatusNotFound, w.Code)
	assert.True(notFound)

	// Test other method than GET (want 308 instead of 301)
	router.PATCH("/path", handlerFunc)
	w = PerformRequest(router, http.MethodPatch, "/path/", nil)

	assert.Equal(http.StatusPermanentRedirect, w.Code)
	assert.Equal("map[Location:[/path]]", fmt.Sprint(w.Header()))

	// Test special case where no node for the prefix "/" exists
	router = Default()
	router.GET("/a", handlerFunc)
	w = PerformRequest(router, http.MethodGet, "/", nil)
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestRouterPanicHandler(t *testing.T) {
	router := Default()
	panicHandled := false

	router.PanicHandler = func(rw http.ResponseWriter, r *http.Request, p interface{}) {
		panicHandled = true
	}

	router.Handle(http.MethodPut, "/user/:name", func(*Context) {
		panic("oops!")
	})

	defer func() {
		if rcv := recover(); rcv != nil {
			t.Fatal("handling panic failed")
		}
	}()

	PerformRequest(router, http.MethodPut, "/user/gopher", nil) // nolint: errcheck
	assert.True(t, panicHandled, "simulating failed")
}

type mockFileSystem struct {
	opened bool
}

func (mfs *mockFileSystem) Open(name string) (http.File, error) {
	mfs.opened = true
	return nil, errors.New("this is just a mock")
}

func TestRouterServeFiles(t *testing.T) {
	router := Default()
	mfs := &mockFileSystem{}

	recv := catchPanic(func() {
		router.ServeFiles("/noFilepath", mfs)
	})
	assert.NotNil(t, recv, "registering path not ending with '*filepath' did not panic")

	router.ServeFiles("/*filepath", mfs)

	PerformRequest(router, http.MethodGet, "/favicon.ico", nil) // nolint: errcheck
	assert.True(t, mfs.opened, "serving file failed")
}
