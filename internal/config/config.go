package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Database
	DB struct {
		Host               string
		Port               string
		User               string
		Password           string
		Name               string
		SSLMode            string
		MaxOpenConns       int
		MaxIdleConns       int
		ConnMaxLifetimeMin int
	}

	// Worker Pool
	WorkerPool struct {
		Size      int
		QueueSize int
	}

	// Server
	Server struct {
		Host            string
		Port            string
		ReadTimeoutSec  int
		WriteTimeoutSec int
		IdleTimeoutSec  int
	}

	// Redis
	// Redis removed (not used)

	// App
	App struct {
		Env      string
		LogLevel string
	}
}

// Load تمام متغیرهای محیطی را بارگذاری می‌کند
func Load() *Config {
	cfg := &Config{}

	// Database
	cfg.DB.Host = getEnv("DB_HOST", "localhost")
	cfg.DB.Port = getEnv("DB_PORT", "5433")
	cfg.DB.User = getEnv("DB_USER", "postgres")
	cfg.DB.Password = getEnv("DB_PASSWORD", "password")
	cfg.DB.Name = getEnv("DB_NAME", "wallet")
	cfg.DB.SSLMode = getEnv("DB_SSLMODE", "disable")
	cfg.DB.MaxOpenConns = getEnvInt("DB_MAX_OPEN_CONNS", 100)
	cfg.DB.MaxIdleConns = getEnvInt("DB_MAX_IDLE_CONNS", 25)
	cfg.DB.ConnMaxLifetimeMin = getEnvInt("DB_CONN_MAX_LIFETIME_MINUTES", 5)

	// Worker Pool
	cfg.WorkerPool.Size = getEnvInt("WORKER_POOL_SIZE", 50)
	cfg.WorkerPool.QueueSize = getEnvInt("WORKER_QUEUE_BUFFER", 100)

	// Server
	cfg.Server.Host = getEnv("SERVER_HOST", "localhost")
	cfg.Server.Port = getEnv("SERVER_PORT", "8080")
	cfg.Server.ReadTimeoutSec = getEnvInt("SERVER_READ_TIMEOUT_SECONDS", 15)
	cfg.Server.WriteTimeoutSec = getEnvInt("SERVER_WRITE_TIMEOUT_SECONDS", 15)
	cfg.Server.IdleTimeoutSec = getEnvInt("SERVER_IDLE_TIMEOUT_SECONDS", 60)

	// Redis removed (not used)

	// App
	cfg.App.Env = getEnv("APP_ENV", "development")
	cfg.App.LogLevel = getEnv("LOG_LEVEL", "debug")

	return cfg
}

// GetDSN اتصال String را برای PostgreSQL می‌سازد
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DB.User,
		c.DB.Password,
		c.DB.Host,
		c.DB.Port,
		c.DB.Name,
		c.DB.SSLMode,
	)
}

// String نمایش کامل Configuration
func (c *Config) String() string {
	var sb strings.Builder
	sb.WriteString("==================================================\n")
	sb.WriteString("📋 Configuration Loaded\n")
	sb.WriteString("==================================================\n")
	sb.WriteString(fmt.Sprintf("Database: %s:%s/%s\n", c.DB.Host, c.DB.Port, c.DB.Name))
	sb.WriteString(fmt.Sprintf("Connection Pool: Max=%d, Idle=%d, Lifetime=%dm\n",
		c.DB.MaxOpenConns, c.DB.MaxIdleConns, c.DB.ConnMaxLifetimeMin))
	sb.WriteString(fmt.Sprintf("Worker Pool: Size=%d, Queue=%d\n", c.WorkerPool.Size, c.WorkerPool.QueueSize))
	sb.WriteString(fmt.Sprintf("Server: %s:%s\n", c.Server.Host, c.Server.Port))
	sb.WriteString(fmt.Sprintf("App Environment: %s (Log: %s)\n", c.App.Env, c.App.LogLevel))
	sb.WriteString("==================================================\n")
	return sb.String()
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
