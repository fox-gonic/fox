package database

import (
	"fmt"
	"net/url"

	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

const (
	maxOpenConns    = 16384
	connMaxLifeTime = 5 // Second
)

// Config mysql config
type Config struct {
	Dialect string `mapstructure:"dialect"`

	// Database name
	Database string `mapstructure:"database"`

	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Encoding string `mapstructure:"encoding"`

	gorm.Config

	*ConnPool
}

// ConnPool config
type ConnPool struct {
	// MaxIdleConns is the maximum number of connections in the idle connection pool
	MaxIdleConns int `mapstructure:"max_idle_conns"`

	// MaxOpenConns is the maximum number of open connections to the database.
	MaxOpenConns int `mapstructure:"max_open_conns"`

	// ConnMaxLifeTime is the maximum amount of time a connection may be reused.
	ConnMaxLifeTime int `mapstructure:"conn_max_life_time"`

	// SetConnMaxIdleTime sets the maximum amount of time a connection may be idle.
	ConnMaxIdleTime int `mapstructure:"conn_max_idle_time"`
}

// Dialector return dialector with config
func (config *Config) Dialector() gorm.Dialector {

	var dialector gorm.Dialector

	if config.Encoding == "" {
		config.Encoding = "utf8mb4"
	}

	switch config.Dialect {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&loc=%s&parseTime=true",
			config.Username,
			config.Password,
			config.Host,
			config.Port,
			config.Database,
			config.Encoding,
			url.QueryEscape("Local"),
		)
		dialector = mysql.Open(dsn)

	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			config.Host,
			config.Username,
			config.Password,
			config.Database,
			config.Port,
		)
		dialector = postgres.Open(dsn)

	case "sqlite":
		dialector = sqlite.Open(config.Database)

	case "sqlserver":
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			config.Username,
			config.Password,
			config.Host,
			config.Port,
			config.Database,
		)
		dialector = sqlserver.Open(dsn)

	case "clickhouse":
		dsn := fmt.Sprintf("tcp://%s:%d?database=%s&username=%s&password=%s&read_timeout=10&write_timeout=20",
			config.Host,
			config.Port,
			config.Database,
			config.Username,
			config.Password,
		)
		dialector = clickhouse.Open(dsn)
	}

	return dialector
}
