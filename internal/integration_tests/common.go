//go:build integration

package integration_tests

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
)

type dbSettings struct {
	user     string
	password string
	host     string
	port     int
	dbname   string
	sslmode  string
	timeout  time.Duration
}

func getEnvVar(varName, defValue string) string {
	setting := os.Getenv(varName)
	if setting == "" {
		setting = defValue
	}
	return setting
}

func newDbSettings() (*dbSettings, error) {
	portStr := getEnvVar("GOABF_DBPORT", "5432")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	return &dbSettings{
		user:     getEnvVar("GOABF_DBUSER", "abfuser"),
		password: getEnvVar("GOABF_DBPASSWORD", "abfpassword"),
		host:     getEnvVar("GOABF_DBHOST", "db"),
		port:     port,
		dbname:   getEnvVar("GOABF_DBNAME", "abf"),
		sslmode:  getEnvVar("GOABF_DBSSLMODE", "disable"),
		timeout:  3 * time.Second,
	}, nil
}

func newDB(s *dbSettings) (*sqlx.DB, error) {
	ctx, _ := context.WithTimeout(context.Background(), s.timeout) //nolint: govet
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		s.user,
		s.password,
		s.host,
		strconv.Itoa(s.port),
		s.dbname,
		s.sslmode,
	)

	return sqlx.ConnectContext(ctx, "pgx", dsn)
}
