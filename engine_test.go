package fox

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEngineAddRoute(t *testing.T) {
	router := New()
	router.addRoute("GET", "/", HandlersChain{func() {}})

	assert.Len(t, router.trees, 1)
	assert.NotNil(t, router.trees["GET"])
	assert.Nil(t, router.trees["POST"])

	router.addRoute("POST", "/", HandlersChain{func() {}})

	assert.Len(t, router.trees, 2)
	assert.NotNil(t, router.trees["GET"])
	assert.NotNil(t, router.trees["POST"])

	router.addRoute("POST", "/post", HandlersChain{func() {}})
	assert.Len(t, router.trees, 2)
}

func TestEngineRegisterRoute(t *testing.T) {
	assert := assert.New(t)
	router := New()

	var index = func(c *Context) string { return "home page" }
	var ping = func() string { return "pong" }

	type HelloHandlerArgs struct {
		Name string `pos:"path:name"`
	}
	var hello = func(c *Context, args *HelloHandlerArgs) {
		fmt.Println("hello", args.Name)
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
		fmt.Println("resource", args.ID)
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
	assert.Empty(w.Body.String())

	w = PerformRequest(router, http.MethodGet, "/group/resources", nil)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal(`["resource1","resource2","resource3"]`, w.Body.String())

	w = PerformRequest(router, http.MethodGet, "/group/resources/1", nil)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal(`{"id":1}`, w.Body.String())
}

func TestEngineRESTful(t *testing.T) {
	assert := assert.New(t)
	router := New()

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
	var create = func(c *Context, args *CreateProductArgs) (*Product, int, error) {
		product := &Product{
			ID:   1,
			Name: args.Name,
			Desc: args.Desc,
		}
		return product, 201, nil
	}

	type GetProductArgs struct {
		ID int `pos:"path:id"`
	}
	var show = func(c *Context, args *GetProductArgs) (*Product, int, error) {
		if args.ID == 0 {
			return nil, 404, nil
		}
		product := &Product{
			ID:   args.ID,
			Name: "Product Name",
			Desc: "Product Desc",
		}
		return product, 200, nil
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
	var destroy = func(c *Context, args *DestroyProductArgs) (*Product, int, error) {
		if args.ID == 0 {
			return nil, 404, nil
		}
		return nil, 200, nil
	}

	router.GET("/products", index)
	router.POST("/products", create)
	router.GET("/products/:id", show)
	router.PATCH("/products/:id", update)
	router.DELETE("/products/:id", destroy)

	w := PerformRequest(router, http.MethodGet, "/products", nil)
	assert.Equal(http.StatusOK, w.Code)
	var response []Product
	json.Unmarshal(w.Body.Bytes(), &response)
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
