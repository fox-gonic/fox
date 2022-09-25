package sessions

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type storeFactory func(*testing.T) Store

const sessionName = "mysession"

const ok = "ok"

func testID(t *testing.T, newStore storeFactory) {
	mux := http.NewServeMux()
	store := newStore(t)
	mux.HandleFunc("/id", func(w http.ResponseWriter, req *http.Request) {
		session := New(sessionName, store, w, req)
		session.Set("key", ok)
		_ = session.Save()
		if session.ID() == "" {
			t.Error("Session id is empty")
		}
		w.WriteHeader(200)
		w.Write([]byte(ok)) // nolint: errcheck
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/id", nil)
	mux.ServeHTTP(res, req)
}

func testGetSet(t *testing.T, newStore storeFactory) {
	mux := http.NewServeMux()
	store := newStore(t)
	mux.HandleFunc("/set", func(w http.ResponseWriter, req *http.Request) {
		session := New(sessionName, store, w, req)
		session.Set("key", ok)
		_ = session.Save()
		w.WriteHeader(200)
		w.Write([]byte(ok)) // nolint: errcheck
	})

	mux.HandleFunc("/get", func(w http.ResponseWriter, req *http.Request) {
		session := New(sessionName, store, w, req)
		if session.Get("key") != ok {
			t.Error("Session writing failed")
		}
		_ = session.Save()
		w.WriteHeader(200)
		w.Write([]byte(ok)) // nolint: errcheck
	})

	res1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/set", nil)
	mux.ServeHTTP(res1, req1)

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/get", nil)
	copyCookies(req2, res1)
	mux.ServeHTTP(res2, req2)
}

func testDeleteKey(t *testing.T, newStore storeFactory) {
	mux := http.NewServeMux()
	store := newStore(t)
	mux.HandleFunc("/set", func(w http.ResponseWriter, req *http.Request) {
		session := New(sessionName, store, w, req)
		session.Set("key", ok)
		_ = session.Save()
		w.WriteHeader(200)
		w.Write([]byte(ok)) // nolint: errcheck
	})

	mux.HandleFunc("/delete", func(w http.ResponseWriter, req *http.Request) {
		session := New(sessionName, store, w, req)
		session.Delete("key")
		_ = session.Save()
		w.WriteHeader(200)
		w.Write([]byte(ok)) // nolint: errcheck
	})

	mux.HandleFunc("/get", func(w http.ResponseWriter, req *http.Request) {
		session := New(sessionName, store, w, req)
		if session.Get("key") != nil {
			t.Error("Session deleting failed")
		}
		_ = session.Save()
		w.WriteHeader(200)
		w.Write([]byte(ok)) // nolint: errcheck
	})

	res1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/set", nil)
	mux.ServeHTTP(res1, req1)

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/delete", nil)
	copyCookies(req2, res1)
	mux.ServeHTTP(res2, req2)

	res3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/get", nil)
	copyCookies(req3, res2)
	mux.ServeHTTP(res3, req3)
}

func testFlashes(t *testing.T, newStore storeFactory) {
	mux := http.NewServeMux()
	store := newStore(t)
	mux.HandleFunc("/set", func(w http.ResponseWriter, req *http.Request) {
		session := New(sessionName, store, w, req)
		session.AddFlash(ok)
		_ = session.Save()
		w.WriteHeader(200)
		w.Write([]byte(ok)) // nolint: errcheck
	})

	mux.HandleFunc("/flash", func(w http.ResponseWriter, req *http.Request) {
		session := New(sessionName, store, w, req)
		l := len(session.Flashes())
		if l != 1 {
			t.Error("Flashes count does not equal 1. Equals ", l)
		}
		_ = session.Save()
		w.WriteHeader(200)
		w.Write([]byte(ok)) // nolint: errcheck
	})

	mux.HandleFunc("/check", func(w http.ResponseWriter, req *http.Request) {
		session := New(sessionName, store, w, req)
		l := len(session.Flashes())
		if l != 0 {
			t.Error("flashes count is not 0 after reading. Equals ", l)
		}
		_ = session.Save()
		w.WriteHeader(200)
		w.Write([]byte(ok)) // nolint: errcheck
	})

	res1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/set", nil)
	mux.ServeHTTP(res1, req1)

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/flash", nil)
	copyCookies(req2, res1)
	mux.ServeHTTP(res2, req2)

	res3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/check", nil)
	copyCookies(req3, res2)
	mux.ServeHTTP(res3, req3)
}

func testClear(t *testing.T, newStore storeFactory) {
	data := map[string]string{
		"key": "val",
		"foo": "bar",
	}
	mux := http.NewServeMux()
	store := newStore(t)
	mux.HandleFunc("/set", func(w http.ResponseWriter, req *http.Request) {
		session := New(sessionName, store, w, req)
		for k, v := range data {
			session.Set(k, v)
		}
		session.Clear()
		_ = session.Save()
		w.WriteHeader(200)
		w.Write([]byte(ok)) // nolint: errcheck
	})

	mux.HandleFunc("/check", func(w http.ResponseWriter, req *http.Request) {
		session := New(sessionName, store, w, req)
		for k, v := range data {
			if session.Get(k) == v {
				t.Fatal("Session clear failed")
			}
		}
		_ = session.Save()
		w.WriteHeader(200)
		w.Write([]byte(ok)) // nolint: errcheck
	})

	res1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/set", nil)
	mux.ServeHTTP(res1, req1)

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/check", nil)
	copyCookies(req2, res1)
	mux.ServeHTTP(res2, req2)
}

func testOptions(t *testing.T, newStore storeFactory) {
	mux := http.NewServeMux()
	store := newStore(t)
	store.Options(Options{
		Domain: "localhost",
	})
	mux.HandleFunc("/domain", func(w http.ResponseWriter, req *http.Request) {
		session := New(sessionName, store, w, req)
		session.Set("key", ok)
		session.Options(Options{
			Path: "/foo/bar/bat",
		})
		_ = session.Save()
		w.WriteHeader(200)
		w.Write([]byte(ok)) // nolint: errcheck
	})
	mux.HandleFunc("/path", func(w http.ResponseWriter, req *http.Request) {
		session := New(sessionName, store, w, req)
		session.Set("key", ok)
		_ = session.Save()
		w.WriteHeader(200)
		w.Write([]byte(ok)) // nolint: errcheck
	})
	mux.HandleFunc("/set", func(w http.ResponseWriter, req *http.Request) {
		session := New(sessionName, store, w, req)
		session.Set("key", ok)
		_ = session.Save()
		w.WriteHeader(200)
		w.Write([]byte(ok)) // nolint: errcheck
	})
	mux.HandleFunc("/expire", func(w http.ResponseWriter, req *http.Request) {
		session := New(sessionName, store, w, req)
		session.Options(Options{
			MaxAge: -1,
		})
		_ = session.Save()
		w.WriteHeader(200)
		w.Write([]byte(ok)) // nolint: errcheck
	})
	mux.HandleFunc("/check", func(w http.ResponseWriter, req *http.Request) {
		session := New(sessionName, store, w, req)
		val := session.Get("key")
		if val != nil {
			t.Fatal("Session expiration failed")
		}
		w.WriteHeader(200)
		w.Write([]byte(ok)) // nolint: errcheck
	})

	testOptionSameSitego(t, mux, newStore)

	res1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/domain", nil)
	mux.ServeHTTP(res1, req1)

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/path", nil)
	mux.ServeHTTP(res2, req2)

	res3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/set", nil)
	mux.ServeHTTP(res3, req3)

	res4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("GET", "/expire", nil)
	mux.ServeHTTP(res4, req4)

	res5 := httptest.NewRecorder()
	req5, _ := http.NewRequest("GET", "/check", nil)
	mux.ServeHTTP(res5, req5)

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

func copyCookies(req *http.Request, res *httptest.ResponseRecorder) {
	req.Header.Set("Cookie", strings.Join(res.Header().Values("Set-Cookie"), "; "))
}

func testOptionSameSitego(t *testing.T, mux *http.ServeMux, newStore storeFactory) {
	mux.HandleFunc("/sameSite", func(w http.ResponseWriter, req *http.Request) {
		session := New(sessionName, newStore(t), w, req)
		session.Set("key", ok)
		session.Options(Options{
			SameSite: http.SameSiteStrictMode,
		})
		_ = session.Save()
		w.WriteHeader(200)
		w.Write([]byte(ok)) // nolint: errcheck
	})

	res3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/sameSite", nil)
	mux.ServeHTTP(res3, req3)

	s := strings.Split(res3.Header().Get("Set-Cookie"), ";")
	if s[1] != " SameSite=Strict" {
		t.Error("Error writing samesite with options:", s[1])
	}
}
