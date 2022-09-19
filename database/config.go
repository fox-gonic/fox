package database

const (
	maxOpenConns    = 16384
	connMaxLifeTime = 5
)

// Config mysql config
type Config struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`

	// Database name
	Database string `mapstructure:"database"`

	// MaxIdleConns is the maximum number of connections in the idle connection pool
	MaxIdleConns int `mapstructure:"max_idle_conns"`

	// MaxOpenConns is the maximum number of open connections to the database.
	MaxOpenConns int `mapstructure:"max_open_conns"`

	// ConnMaxLifeTime is the maximum amount of time a connection may be reused.
	ConnMaxLifeTime int `mapstructure:"conn_max_life_time"`
}
