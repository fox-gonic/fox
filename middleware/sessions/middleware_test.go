package sessions

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"strings"
// 	"testing"

// 	"gorm.io/driver/sqlite"
// 	"gorm.io/gorm"

// 	"github.com/fox-gonic/fox/middleware/sessions"
// )

// type storeFactory func(*testing.T) sessions.Store

// const sessionName = "mysession"

// const ok = "ok"

// func testSessionID(t *testing.T, newStore storeFactory) {
// 	r := Default()
// 	r.Use(NewSessionHandler(sessionName, newStore(t)))

// 	r.GET("/id", func(c *Context) string {
// 		session := DefaultSession(c)
// 		session.Set("key", ok)
// 		_ = session.Save()
// 		if session.ID() == "" {
// 			t.Error("Session id is empty")
// 		}
// 		return ok
// 	})

// 	res1 := httptest.NewRecorder()
// 	req1, _ := http.NewRequest("GET", "/id", nil)
// 	r.ServeHTTP(res1, req1)
// }

// func testSessionGetSet(t *testing.T, newStore storeFactory) {
// 	r := Default()
// 	r.Use(NewSessionHandler(sessionName, newStore(t)))

// 	r.GET("/set", func(c *Context) string {
// 		session := DefaultSession(c)
// 		session.Set("key", ok)
// 		_ = session.Save()
// 		return ok
// 	})

// 	r.GET("/get", func(c *Context) string {
// 		session := DefaultSession(c)
// 		if session.Get("key") != ok {
// 			t.Error("Session writing failed")
// 		}
// 		_ = session.Save()
// 		return ok
// 	})

// 	res1 := httptest.NewRecorder()
// 	req1, _ := http.NewRequest("GET", "/set", nil)
// 	r.ServeHTTP(res1, req1)

// 	res2 := httptest.NewRecorder()
// 	req2, _ := http.NewRequest("GET", "/get", nil)
// 	testSessionCopyCookies(req2, res1)
// 	r.ServeHTTP(res2, req2)
// }

// func testSessionDeleteKey(t *testing.T, newStore storeFactory) {
// 	r := Default()
// 	r.Use(NewSessionHandler(sessionName, newStore(t)))

// 	r.GET("/set", func(c *Context) string {
// 		session := DefaultSession(c)
// 		session.Set("key", ok)
// 		_ = session.Save()
// 		return ok
// 	})

// 	r.GET("/delete", func(c *Context) string {
// 		session := DefaultSession(c)
// 		session.Delete("key")
// 		_ = session.Save()
// 		return ok
// 	})

// 	r.GET("/get", func(c *Context) string {
// 		session := DefaultSession(c)
// 		if session.Get("key") != nil {
// 			t.Error("Session deleting failed")
// 		}
// 		_ = session.Save()
// 		return ok
// 	})

// 	res1 := httptest.NewRecorder()
// 	req1, _ := http.NewRequest("GET", "/set", nil)
// 	r.ServeHTTP(res1, req1)

// 	res2 := httptest.NewRecorder()
// 	req2, _ := http.NewRequest("GET", "/delete", nil)
// 	testSessionCopyCookies(req2, res1)
// 	r.ServeHTTP(res2, req2)

// 	res3 := httptest.NewRecorder()
// 	req3, _ := http.NewRequest("GET", "/get", nil)
// 	testSessionCopyCookies(req3, res2)
// 	r.ServeHTTP(res3, req3)
// }

// func testSessionFlashes(t *testing.T, newStore storeFactory) {
// 	r := Default()
// 	store := newStore(t)
// 	r.Use(NewSessionHandler(sessionName, store))

// 	r.GET("/set", func(c *Context) string {
// 		session := DefaultSession(c)
// 		session.AddFlash(ok)
// 		_ = session.Save()
// 		return ok
// 	})

// 	r.GET("/flash", func(c *Context) string {
// 		session := DefaultSession(c)
// 		l := len(session.Flashes())
// 		if l != 1 {
// 			t.Error("Flashes count does not equal 1. Equals ", l)
// 		}
// 		_ = session.Save()
// 		return ok
// 	})

// 	r.GET("/check", func(c *Context) string {
// 		session := DefaultSession(c)
// 		l := len(session.Flashes())
// 		if l != 0 {
// 			t.Error("flashes count is not 0 after reading. Equals ", l)
// 		}
// 		_ = session.Save()
// 		return ok
// 	})

// 	res1 := httptest.NewRecorder()
// 	req1, _ := http.NewRequest("GET", "/set", nil)
// 	r.ServeHTTP(res1, req1)

// 	res2 := httptest.NewRecorder()
// 	req2, _ := http.NewRequest("GET", "/flash", nil)
// 	testSessionCopyCookies(req2, res1)
// 	r.ServeHTTP(res2, req2)

// 	res3 := httptest.NewRecorder()
// 	req3, _ := http.NewRequest("GET", "/check", nil)
// 	testSessionCopyCookies(req3, res2)
// 	r.ServeHTTP(res3, req3)
// }

// func testSessionClear(t *testing.T, newStore storeFactory) {
// 	data := map[string]string{
// 		"key": "val",
// 		"foo": "bar",
// 	}
// 	r := Default()
// 	store := newStore(t)
// 	r.Use(NewSessionHandler(sessionName, store))

// 	r.GET("/set", func(c *Context) string {
// 		session := DefaultSession(c)
// 		for k, v := range data {
// 			session.Set(k, v)
// 		}
// 		session.Clear()
// 		_ = session.Save()
// 		return ok
// 	})

// 	r.GET("/check", func(c *Context) string {
// 		session := DefaultSession(c)
// 		for k, v := range data {
// 			if session.Get(k) == v {
// 				t.Fatal("Session clear failed")
// 			}
// 		}
// 		_ = session.Save()
// 		return ok
// 	})

// 	res1 := httptest.NewRecorder()
// 	req1, _ := http.NewRequest("GET", "/set", nil)
// 	r.ServeHTTP(res1, req1)

// 	res2 := httptest.NewRecorder()
// 	req2, _ := http.NewRequest("GET", "/check", nil)
// 	testSessionCopyCookies(req2, res1)
// 	r.ServeHTTP(res2, req2)
// }

// func testSessionOptions(t *testing.T, newStore storeFactory) {
// 	r := Default()
// 	store := newStore(t)
// 	store.Options(sessions.Options{
// 		Domain: "localhost",
// 	})
// 	r.Use(NewSessionHandler(sessionName, store))

// 	r.GET("/domain", func(c *Context) string {
// 		session := DefaultSession(c)
// 		session.Set("key", ok)
// 		session.Options(sessions.Options{
// 			Path: "/foo/bar/bat",
// 		})
// 		_ = session.Save()
// 		return ok
// 	})
// 	r.GET("/path", func(c *Context) string {
// 		session := DefaultSession(c)
// 		session.Set("key", ok)
// 		_ = session.Save()
// 		return ok
// 	})
// 	r.GET("/set", func(c *Context) string {
// 		session := DefaultSession(c)
// 		session.Set("key", ok)
// 		_ = session.Save()
// 		return ok
// 	})
// 	r.GET("/expire", func(c *Context) string {
// 		session := DefaultSession(c)
// 		session.Options(sessions.Options{
// 			MaxAge: -1,
// 		})
// 		_ = session.Save()
// 		return ok
// 	})
// 	r.GET("/check", func(c *Context) string {
// 		session := DefaultSession(c)
// 		val := session.Get("key")
// 		if val != nil {
// 			t.Fatal("Session expiration failed")
// 		}
// 		return ok
// 	})

// 	testSessionOptionSameSitego(t, r)

// 	res1 := httptest.NewRecorder()
// 	req1, _ := http.NewRequest("GET", "/domain", nil)
// 	r.ServeHTTP(res1, req1)

// 	res2 := httptest.NewRecorder()
// 	req2, _ := http.NewRequest("GET", "/path", nil)
// 	r.ServeHTTP(res2, req2)

// 	res3 := httptest.NewRecorder()
// 	req3, _ := http.NewRequest("GET", "/set", nil)
// 	r.ServeHTTP(res3, req3)

// 	res4 := httptest.NewRecorder()
// 	req4, _ := http.NewRequest("GET", "/expire", nil)
// 	r.ServeHTTP(res4, req4)

// 	res5 := httptest.NewRecorder()
// 	req5, _ := http.NewRequest("GET", "/check", nil)
// 	r.ServeHTTP(res5, req5)

// 	for _, c := range res1.Header().Values("Set-Cookie") {
// 		s := strings.Split(c, ";")
// 		if s[1] != " Path=/foo/bar/bat" {
// 			t.Error("Error writing path with options:", s[1])
// 		}
// 	}

// 	for _, c := range res2.Header().Values("Set-Cookie") {
// 		s := strings.Split(c, ";")
// 		if s[1] != " Domain=localhost" {
// 			t.Error("Error writing domain with options:", s[1])
// 		}
// 	}
// }

// func testSessionMany(t *testing.T, newStore storeFactory) {
// 	r := Default()
// 	sessionNames := []string{"a", "b"}

// 	r.Use(ManySessionHandler(sessionNames, newStore(t)))

// 	r.GET("/set", func(c *Context) string {
// 		sessionA := DefaultMany(c, "a")
// 		sessionA.Set("hello", "world")
// 		_ = sessionA.Save()

// 		sessionB := DefaultMany(c, "b")
// 		sessionB.Set("foo", "bar")
// 		_ = sessionB.Save()
// 		return ok
// 	})

// 	r.GET("/get", func(c *Context) string {
// 		sessionA := DefaultMany(c, "a")
// 		if sessionA.Get("hello") != "world" {
// 			t.Error("Session writing failed")
// 		}
// 		_ = sessionA.Save()

// 		sessionB := DefaultMany(c, "b")
// 		if sessionB.Get("foo") != "bar" {
// 			t.Error("Session writing failed")
// 		}
// 		_ = sessionB.Save()
// 		return ok
// 	})

// 	res1 := httptest.NewRecorder()
// 	req1, _ := http.NewRequest("GET", "/set", nil)
// 	r.ServeHTTP(res1, req1)

// 	res2 := httptest.NewRecorder()
// 	req2, _ := http.NewRequest("GET", "/get", nil)
// 	header := ""
// 	for _, x := range res1.Header()["Set-Cookie"] {
// 		header += strings.Split(x, ";")[0] + "; \n"
// 	}
// 	req2.Header.Set("Cookie", header)
// 	r.ServeHTTP(res2, req2)
// }

// func testSessionCopyCookies(req *http.Request, res *httptest.ResponseRecorder) {
// 	req.Header.Set("Cookie", strings.Join(res.Header().Values("Set-Cookie"), "; "))
// }

// func testSessionOptionSameSitego(t *testing.T, r *Engine) {
// 	r.GET("/sameSite", func(c *Context) string {
// 		session := DefaultSession(c)
// 		session.Set("key", ok)
// 		session.Options(sessions.Options{
// 			SameSite: http.SameSiteStrictMode,
// 		})
// 		_ = session.Save()
// 		return ok
// 	})

// 	res3 := httptest.NewRecorder()
// 	req3, _ := http.NewRequest("GET", "/sameSite", nil)
// 	r.ServeHTTP(res3, req3)

// 	s := strings.Split(res3.Header().Get("Set-Cookie"), ";")
// 	if s[1] != " SameSite=Strict" {
// 		t.Error("Error writing samesite with options:", s[1])
// 	}
// }

// // test cookie store ----------------------------------------------------------

// var testCookieStore = func(_ *testing.T) sessions.Store {
// 	store := sessions.NewCookieStore([]byte("secret"))
// 	return store
// }

// func TestCookieSessionGetSet(t *testing.T) {
// 	testSessionGetSet(t, testCookieStore)
// }

// func TestCookieSessionDeleteKey(t *testing.T) {
// 	testSessionDeleteKey(t, testCookieStore)
// }

// func TestCookieSessionFlashes(t *testing.T) {
// 	testSessionFlashes(t, testCookieStore)
// }

// func TestCookieSessionClear(t *testing.T) {
// 	testSessionClear(t, testCookieStore)
// }

// func TestCookieSessionOptions(t *testing.T) {
// 	testSessionOptions(t, testCookieStore)
// }

// func TestCookieSessionMany(t *testing.T) {
// 	testSessionMany(t, testCookieStore)
// }

// // test gorm store ------------------------------------------------------------

// var testGormStore = func(_ *testing.T) sessions.Store {
// 	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
// 	if err != nil {
// 		panic(err)
// 	}
// 	return sessions.NewGormStore(db, true, []byte("secret"))
// }

// func TestGormSessionID(t *testing.T) {
// 	testSessionID(t, testGormStore)
// }

// func TestGormSessionGetSet(t *testing.T) {
// 	testSessionGetSet(t, testGormStore)
// }

// func TestGormSessionDeleteKey(t *testing.T) {
// 	testSessionDeleteKey(t, testGormStore)
// }

// func TestGormSessionFlashes(t *testing.T) {
// 	testSessionFlashes(t, testGormStore)
// }

// func TestGormSessionClear(t *testing.T) {
// 	testSessionClear(t, testGormStore)
// }

// func TestGormSessionOptions(t *testing.T) {
// 	testSessionOptions(t, testGormStore)
// }

// func TestGormSessionMany(t *testing.T) {
// 	testSessionMany(t, testGormStore)
// }
