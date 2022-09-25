package sessions

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var testGormStore = func(_ *testing.T) Store {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return NewGormStore(db, true, []byte("secret"))
}

func TestGorm_SessionID(t *testing.T) {
	testID(t, testGormStore)
}

func TestGorm_SessionGetSet(t *testing.T) {
	testGetSet(t, testGormStore)
}

func TestGorm_SessionDeleteKey(t *testing.T) {
	testDeleteKey(t, testGormStore)
}

func TestGorm_SessionFlashes(t *testing.T) {
	testFlashes(t, testGormStore)
}

func TestGorm_SessionClear(t *testing.T) {
	testClear(t, testGormStore)
}

func TestGorm_SessionOptions(t *testing.T) {
	testOptions(t, testGormStore)
}
