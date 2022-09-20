// Package database https://gorm.io/zh_CN/docs/connecting_to_the_database.html
package database

import (
	"time"

	"gorm.io/gorm"

	log "github.com/fox-gonic/fox/logger"
)

// Database instance type
type Database struct {
	*gorm.DB
}

// NowFunc returns current time, this function is exported in order to be able
// to give the flexibility to the developer to customize it according to their needs
var NowFunc = func() time.Time {
	return time.Now().UTC()
}

// New database with configuration
func New(config *Config) (database *Database, err error) {

	var dialector = config.Dialector()

	database, err = NewWithDialector(dialector, &config.Config)
	if err != nil {
		return nil, err
	}

	if config.ConnPool != nil {
		err = database.SetConnPool(*config.ConnPool)
	}
	return
}

// NewWithDialector database with dialector
func NewWithDialector(dialector gorm.Dialector, config *gorm.Config) (database *Database, err error) {

	if config == nil {
		config = &gorm.Config{}
	}

	config.DisableForeignKeyConstraintWhenMigrating = true

	if config.NowFunc == nil {
		config.NowFunc = NowFunc
	}

	if config.Logger == nil {
		config.Logger = NewLogger(50)
	}

	db, err := gorm.Open(dialector, config)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	database = &Database{
		DB: db,
	}
	return
}

// SetConnPool set connection pool
func (database *Database) SetConnPool(pool ConnPool) error {
	sqlDB, err := database.DB.DB()
	if err != nil {
		return err
	}

	if pool.MaxOpenConns > 0 {
		if pool.MaxOpenConns > maxOpenConns {
			pool.MaxOpenConns = maxOpenConns
		}

		// SetMaxOpenConns sets the maximum number of open connections to the database.
		sqlDB.SetMaxOpenConns(pool.MaxOpenConns)
	}

	if pool.MaxIdleConns > 0 {
		if pool.MaxIdleConns > pool.MaxOpenConns {
			pool.MaxIdleConns = pool.MaxOpenConns
		}

		// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
		sqlDB.SetMaxIdleConns(pool.MaxIdleConns)
	}

	// set op response timeout
	if pool.ConnMaxLifeTime == 0 {
		pool.ConnMaxLifeTime = connMaxLifeTime
	}

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Duration(pool.ConnMaxLifeTime) * time.Second)

	// SetConnMaxIdleTime sets the maximum amount of time a connection may be idle.
	sqlDB.SetConnMaxIdleTime(time.Duration(pool.ConnMaxIdleTime) * time.Second)

	return nil
}

// Get gorm.DB instance with request id
func (database *Database) Get(requestID ...string) *gorm.DB {

	var traceID string
	if len(requestID) > 0 {
		traceID = requestID[0]
	} else {
		traceID = log.DefaultGenRequestID()
	}

	// create new database session
	db := database.Session(&gorm.Session{
		Logger: NewLogger(0, traceID),
	})

	return db
}
