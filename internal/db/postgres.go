package db

import (
	"fmt"
	"sync"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PostgresConfig persists the config for our PostgreSQL database connection
type PostgresConfig struct {
	Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
	User     string `env:"POSTGRES_USER"`
	Password string `env:"POSTGRES_PASSWORD"`
	Database string `env:"POSTGRES_DB"`
}

// PostgresSuperUser persists the config for our PostgreSQL superuser
type PostgresSuperUser struct {
	Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
	User     string `env:"POSTGRES_SUPERUSER" envDefault:"postgres"`
	Password string `env:"POSTGRES_SUPERUSER_PASSWORD" envDefault:""`
	Database string `env:"POSTGRES_SUPERUSER_DB" envDefault:"postgres"`
}

// DBClient manages the database connection
type DBClient struct {
	db *sqlx.DB
}

// Global singleton instance
var dbClient *DBClient
var once sync.Once

// GetDBClient returns a singleton database client
func GetDBClient() *DBClient {
	once.Do(func() {
		db := openDBConnection()
		dbClient = &DBClient{db: db}
	})
	return dbClient
}

// GetConnection returns the database connection
// usage: db := db.GetDBClient().GetConnection()
func (c *DBClient) GetConnection() *sqlx.DB {
	return c.db
}

// Close closes the database connection
// usage: defer db.GetDBClient().Close()
func (c *DBClient) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// openDBConnection creates a new database connection
func openDBConnection() *sqlx.DB {
	c := getPostgresConfig()
	db, err := sqlx.Connect("postgres", buildPostgresURL(c))
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}

	err = db.Ping()
	if err != nil {
		logger.Fatal("Failed to ping database", "error", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	logger.Info("Successfully connected to database")
	return db
}

// GetPostgresConfig returns a PostgresConfig pointer with the correct Postgres Config values
func getPostgresConfig() *PostgresConfig {
	c := PostgresConfig{}
	if err := env.Parse(&c); err != nil {
		fmt.Printf("%+v\n", err)
	}
	return &c
}

// buildPostgresURL builds a Postgres URL from a PostgresConfig
func buildPostgresURL(c *PostgresConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", c.User, c.Password, c.Host, c.Port, c.Database)
}
