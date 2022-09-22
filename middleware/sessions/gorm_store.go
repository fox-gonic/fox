package sessions

import (
	"time"

	"github.com/wader/gormstore/v2"
	"gorm.io/gorm"
)

// NewGormStore creates a new gormstore session
func NewGormStore(d *gorm.DB, expiredSessionCleanup bool, keyPairs ...[]byte) Store {
	s := gormstore.New(d, keyPairs...)
	if expiredSessionCleanup {
		quit := make(chan struct{})
		go s.PeriodicCleanup(1*time.Hour, quit)
	}
	return &GormStore{s}
}

// GormStore represent a gormstore
type GormStore struct {
	*gormstore.Store
}

// Options implemented Store
func (s *GormStore) Options(options Options) {
	s.Store.SessionOpts = &options
}
