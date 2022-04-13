package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	MongoDBDriverDefault = "mongodb"
	MongoDBDriverAtlas   = "mongodb_atlas"
	MongoDBSchemeDefault = "mongodb"
	MongoDBSchemeAtlas   = "mongodb+srv"
)

type Config struct {
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	DBName   string `yaml:"dbName"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"sslMode"`

	MaxOpenConnections    int `yaml:"maxOpenConnections"`
	MaxLifeTimeConnection int `yaml:"maxLifeTimeConnection"` // Seconds
	MaxIdleConnections    int `yaml:"maxIdleConnections"`
	MaxIdleTimeConnection int `yaml:"maxIdleTimeConnection"` // Seconds
}

// NewDB create new DB connection.
func NewDB(config *Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	conn, err := sqlx.Open(config.Driver, dsn)
	if err != nil {
		return nil, err
	}

	conn.SetConnMaxLifetime(time.Duration(config.MaxLifeTimeConnection) * time.Second)
	conn.SetMaxOpenConns(config.MaxOpenConnections)
	conn.SetMaxIdleConns(config.MaxIdleConnections)
	conn.SetConnMaxIdleTime(time.Duration(config.MaxIdleTimeConnection) * time.Second)

	if er := conn.Ping(); er != nil {
		return nil, er
	}

	return conn, nil
}

// NewMongo create new mongodb connection.
func NewMongo(config *Config) (*mongo.Database, error) {
	if config.Driver == "" {
		config.Driver = MongoDBDriverDefault
	}

	scheme := MongoDBSchemeDefault
	if config.Driver == MongoDBDriverAtlas {
		scheme = MongoDBSchemeAtlas
	}

	dsn := fmt.Sprintf(
		"%s://%s:%s@%s/%s?retryWrites=true&w=majority",
		scheme, config.User, config.Password, config.Host, config.DBName,
	)

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(dsn))
	if err != nil {
		return nil, err
	}

	if er := client.Ping(context.TODO(), nil); er != nil {
		return nil, er
	}

	return client.Database(config.DBName), nil
}
