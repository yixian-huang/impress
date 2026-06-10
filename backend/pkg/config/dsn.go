package config

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// PostgresInput holds PostgreSQL connection fields from the setup wizard.
type PostgresInput struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

// DatabaseInput is the database section of setup bootstrap requests.
type DatabaseInput struct {
	Type       string         `json:"type"`
	SQLitePath string         `json:"sqlitePath"`
	Postgres   *PostgresInput `json:"postgres"`
}

// BuildDSN converts wizard database input into a driver DSN string.
func BuildDSN(in DatabaseInput) (string, error) {
	dbType := strings.ToLower(strings.TrimSpace(in.Type))
	switch dbType {
	case "", "sqlite":
		path := strings.TrimSpace(in.SQLitePath)
		if path == "" {
			path = "./data/impress.db"
		}
		if strings.HasPrefix(path, ":memory:") {
			return path, nil
		}
		if !strings.HasPrefix(path, "file:") {
			path = "file:" + path
		}
		if !strings.Contains(path, "?") {
			path += "?cache=shared&mode=rwc"
		}
		return path, nil
	case "postgres", "postgresql":
		if in.Postgres == nil {
			return "", fmt.Errorf("postgres connection details are required")
		}
		return buildPostgresDSN(*in.Postgres)
	default:
		return "", fmt.Errorf("unsupported database type %q", in.Type)
	}
}

func buildPostgresDSN(pg PostgresInput) (string, error) {
	host := strings.TrimSpace(pg.Host)
	user := strings.TrimSpace(pg.User)
	dbname := strings.TrimSpace(pg.DBName)
	if host == "" || user == "" || dbname == "" {
		return "", fmt.Errorf("postgres host, user, and dbname are required")
	}
	port := pg.Port
	if port == 0 {
		port = 5432
	}
	sslmode := strings.TrimSpace(pg.SSLMode)
	if sslmode == "" {
		sslmode = "disable"
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pg.Password),
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   "/" + dbname,
	}
	q := u.Query()
	q.Set("sslmode", sslmode)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// DatabaseTypeFromDSN returns a short label for API responses.
func DatabaseTypeFromDSN(dsn string) string {
	if isPostgresDSN(dsn) {
		return "postgres"
	}
	return "sqlite"
}

func isPostgresDSN(dsn string) bool {
	normalized := strings.ToLower(strings.TrimSpace(dsn))
	if strings.HasPrefix(normalized, "postgres://") || strings.HasPrefix(normalized, "postgresql://") {
		return true
	}
	signals := 0
	for _, key := range []string{"host=", "dbname=", "user=", "password=", "sslmode="} {
		if strings.Contains(normalized, key) {
			signals++
		}
	}
	return signals >= 2
}

// FormatPostgresPort returns a normalized postgres port.
func FormatPostgresPort(port int) int {
	if port <= 0 {
		return 5432
	}
	return port
}

// PostgresPortString formats port for libpq-style DSN if needed later.
func PostgresPortString(port int) string {
	return strconv.Itoa(FormatPostgresPort(port))
}
