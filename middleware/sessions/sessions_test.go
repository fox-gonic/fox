package sessions

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fox-gonic/fox"
)

type storeFactory func(*testing.T) Store

const sessionName = "mysession"

const ok = "ok"

func testGetSet(t *testing.T, newStore storeFactory) {
	r := fox.New()
	r.Use(New(sessionName, newStore(t)))

	r.GET("/set", func(c *fox.Context) string {
		session := Default(c)
		session.Set("key", ok)
		_ = session.Save()
		return ok
	})

	r.GET("/get", func(c *fox.Context) string {
		session := Default(c)
		if session.Get("key") != ok {
			t.Error("Session writing failed")
		}
		_ = session.Save()
		return ok
	})

	res1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/set", nil)
	r.ServeHTTP(res1, req1)

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/get", nil)
	copyCookies(req2, res1)
	r.ServeHTTP(res2, req2)
}

func testDeleteKey(t *testing.T, newStore storeFactory) {
	r := fox.New()
	r.Use(New(sessionName, newStore(t)))

	r.GET("/set", func(c *fox.Context) string {
		session := Default(c)
		session.Set("key", ok)
		_ = session.Save()
		return ok
	})

	r.GET("/delete", func(c *fox.Context) string {
		session := Default(c)
		session.Delete("key")
		_ = session.Save()
		return ok
	})

	r.GET("/get", func(c *fox.Context) string {
		session := Default(c)
		if session.Get("key") != nil {
			t.Error("Session deleting failed")
		}
		_ = session.Save()
		return ok
	})

	res1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/set", nil)
	r.ServeHTTP(res1, req1)

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/delete", nil)
	copyCookies(req2, res1)
	r.ServeHTTP(res2, req2)

	res3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/get", nil)
	copyCookies(req3, res2)
	r.ServeHTTP(res3, req3)
}

func test(t *testing.T, newStore storeFactory) {
	r := fox.New()
	store := newStore(t)
	r.Use(New(sessionName, store))

	r.GET("/set", func(c *fox.Context) string {
		session := Default(c)
		session.AddFlash(ok)
		_ = session.Save()
		return ok
	})

	r.GET("/flash", func(c *fox.Context) string {
		session := Default(c)
		l := len(session.Flashes())
		if l != 1 {
			t.Error("Flashes count does not equal 1. Equals ", l)
		}
		_ = session.Save()
		return ok
	})

	r.GET("/check", func(c *fox.Context) string {
		session := Default(c)
		l := len(session.Flashes())
		if l != 0 {
			t.Error("flashes count is not 0 after reading. Equals ", l)
		}
		_ = session.Save()
		return ok
	})

	res1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/set", nil)
	r.ServeHTTP(res1, req1)

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/flash", nil)
	copyCookies(req2, res1)
	r.ServeHTTP(res2, req2)

	res3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/check", nil)
	copyCookies(req3, res2)
	r.ServeHTTP(res3, req3)
}

func testClear(t *testing.T, newStore storeFactory) {
	data := map[string]string{
		"key": "val",
		"foo": "bar",
	}
	r := fox.New()
	store := newStore(t)
	r.Use(New(sessionName, store))

	r.GET("/set", func(c *fox.Context) string {
		session := Default(c)
		for k, v := range data {
			session.Set(k, v)
		}
		session.Clear()
		_ = session.Save()
		return ok
	})

	r.GET("/check", func(c *fox.Context) string {
		session := Default(c)
		for k, v := range data {
			if session.Get(k) == v {
				t.Fatal("Session clear failed")
			}
		}
		_ = session.Save()
		return ok
	})

	res1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/set", nil)
	r.ServeHTTP(res1, req1)

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/check", nil)
	copyCookies(req2, res1)
	r.ServeHTTP(res2, req2)
}

func testOptions(t *testing.T, newStore storeFactory) {
	r := fox.New()
	store := newStore(t)
	store.Options(Options{
		Domain: "localhost",
	})
	r.Use(New(sessionName, store))

	r.GET("/domain", func(c *fox.Context) string {
		session := Default(c)
		session.Set("key", ok)
		session.Options(Options{
			Path: "/foo/bar/bat",
		})
		_ = session.Save()
		return ok
	})
	r.GET("/path", func(c *fox.Context) string {
		session := Default(c)
		session.Set("key", ok)
		_ = session.Save()
		return ok
	})
	r.GET("/set", func(c *fox.Context) string {
		session := Default(c)
		session.Set("key", ok)
		_ = session.Save()
		return ok
	})
	r.GET("/expire", func(c *fox.Context) string {
		session := Default(c)
		session.Options(Options{
			MaxAge: -1,
		})
		_ = session.Save()
		return ok
	})
	r.GET("/check", func(c *fox.Context) string {
		session := Default(c)
		val := session.Get("key")
		if val != nil {
			t.Fatal("Session expiration failed")
		}
		return ok
	})

	testOptionSameSitego(t, r)

	res1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/domain", nil)
	r.ServeHTTP(res1, req1)

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/path", nil)
	r.ServeHTTP(res2, req2)

	res3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/set", nil)
	r.ServeHTTP(res3, req3)

	res4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("GET", "/expire", nil)
	r.ServeHTTP(res4, req4)

	res5 := httptest.NewRecorder()
	req5, _ := http.NewRequest("GET", "/check", nil)
	r.ServeHTTP(res5, req5)

	for _, c := range res1.Header().Values("Set-Cookie") {
		s := strings.Split(c, ";")
		if s[1] != " Path=/foo/bar/bat" {
			t.Error("Error writing path with options:", s[1])
		}
	}

	for _, c := range res2.Header().Values("Set-Cookie") {
		s := strings.Split(c, ";")
		if s[1] != " Domain=localhost" {
			t.Error("Error writing domain with options:", s[1])
		}
	}
}

func testMany(t *testing.T, newStore storeFactory) {
	r := fox.New()
	sessionNames := []string{"a", "b"}

	r.Use(Many(sessionNames, newStore(t)))

	r.GET("/set", func(c *fox.Context) string {
		sessionA := DefaultMany(c, "a")
		sessionA.Set("hello", "world")
		_ = sessionA.Save()

		sessionB := DefaultMany(c, "b")
		sessionB.Set("foo", "bar")
		_ = sessionB.Save()
		return ok
	})

	r.GET("/get", func(c *fox.Context) string {
		sessionA := DefaultMany(c, "a")
		if sessionA.Get("hello") != "world" {
			t.Error("Session writing failed")
		}
		_ = sessionA.Save()

		sessionB := DefaultMany(c, "b")
		if sessionB.Get("foo") != "bar" {
			t.Error("Session writing failed")
		}
		_ = sessionB.Save()
		return ok
	})

	res1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/set", nil)
	r.ServeHTTP(res1, req1)

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/get", nil)
	header := ""
	for _, x := range res1.Header()["Set-Cookie"] {
		header += strings.Split(x, ";")[0] + "; \n"
	}
	req2.Header.Set("Cookie", header)
	r.ServeHTTP(res2, req2)
}

func copyCookies(req *http.Request, res *httptest.ResponseRecorder) {
	req.Header.Set("Cookie", strings.Join(res.Header().Values("Set-Cookie"), "; "))
}

func testOptionSameSitego(t *testing.T, r *fox.Engine) {
	r.GET("/sameSite", func(c *fox.Context) string {
		session := Default(c)
		session.Set("key", ok)
		session.Options(Options{
			SameSite: http.SameSiteStrictMode,
		})
		_ = session.Save()
		return ok
	})

	res3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/sameSite", nil)
	r.ServeHTTP(res3, req3)

	s := strings.Split(res3.Header().Get("Set-Cookie"), ";")
	if s[1] != " SameSite=Strict" {
		t.Error("Error writing samesite with options:", s[1])
	}
}
