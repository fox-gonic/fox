// Package database https://gorm.io/zh_CN/docs/connecting_to_the_database.html
package database

import (
	"fmt"
	"net/url"
	"time"

	"github.com/fox-gonic/fox/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Dialector return dialector with config
func (config *Config) Dialector() gorm.Dialector {

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&loc=%s&parseTime=true",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		url.QueryEscape("Local"),
	)

	return mysql.Open(dsn)
}

// Database instance type
type Database struct {
	*gorm.DB
}

// NowFunc returns current time, this function is exported in order to be able
// to give the flexibility to the developer to customize it according to their needs
var NowFunc = func() time.Time {
	return time.Now().UTC()
}

// NewDatabase database with configuration
func NewDatabase(config *Config) (database *Database, err error) {

	var dialector = config.Dialector()

	db, err := gorm.Open(dialector, &gorm.Config{
		NowFunc:                                  NowFunc,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
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

	// * set connection pool
	// ******************************************************************

	if config.MaxOpenConns > 0 {
		if config.MaxOpenConns > maxOpenConns {
			config.MaxOpenConns = maxOpenConns
		}

		// SetMaxOpenConns sets the maximum number of open connections to the database.
		sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	}

	if config.MaxIdleConns > 0 {
		if config.MaxIdleConns > config.MaxOpenConns {
			config.MaxIdleConns = config.MaxOpenConns
		}

		// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
		sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	}

	// set op response timeout
	if config.ConnMaxLifeTime == 0 {
		config.ConnMaxLifeTime = connMaxLifeTime
	}

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Duration(config.ConnMaxLifeTime) * time.Second)

	database = &Database{
		DB: db,
	}

	return
}

// Get gorm.DB instance with xReqID
func (database *Database) Get(requestID ...string) *gorm.DB {

	var traceID string
	if len(requestID) > 0 {
		traceID = requestID[0]
	} else {
		traceID = logger.DefaultGenRequestID()
	}

	// create new database session
	db := database.Session(&gorm.Session{
		Logger: newLog(0, traceID),
	})

	return db
}
