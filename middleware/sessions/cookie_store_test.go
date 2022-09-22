package sessions

import (
	"testing"
)

var testCookieStore = func(_ *testing.T) Store {
	store := NewCookieStore([]byte("secret"))
	return store
}

func TestCookie_SessionGetSet(t *testing.T) {
	testGetSet(t, testCookieStore)
}

func TestCookie_SessionDeleteKey(t *testing.T) {
	testDeleteKey(t, testCookieStore)
}

func TestCookie_SessionFlashes(t *testing.T) {
	test(t, testCookieStore)
}

func TestCookie_SessionClear(t *testing.T) {
	testClear(t, testCookieStore)
}

func TestCookie_SessionOptions(t *testing.T) {
	testOptions(t, testCookieStore)
}

func TestCookie_SessionMany(t *testing.T) {
	testMany(t, testCookieStore)
}
