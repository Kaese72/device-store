package mariadb

import (
	"context"
	"database/sql"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/Kaese72/device-store/internal/config"
	tcmariadb "github.com/testcontainers/testcontainers-go/modules/mariadb"

	// NewMariadbPersistence opens the "mysql" driver name registered by this
	// package's init(); production only pulls it in via main.go, so tests
	// need the same blank import to register the driver themselves.
	_ "go.elastic.co/apm/module/apmsql/mysql"
)

// setupTestDB starts a throwaway MariaDB container, applies every migration
// in migrations/ against it (using the real SQL files, not a hand-maintained
// copy), and returns a mariadbPersistence connected to it. The container is
// torn down automatically when the test completes.
func setupTestDB(t *testing.T) mariadbPersistence {
	t.Helper()
	ctx := context.Background()

	container, err := tcmariadb.Run(ctx, "mariadb:11",
		tcmariadb.WithDatabase("devicestore"),
		tcmariadb.WithUsername("devicestore"),
		tcmariadb.WithPassword("devicestore"),
	)
	if err != nil {
		t.Fatalf("failed to start mariadb container: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Logf("failed to terminate mariadb container: %v", err)
		}
	})

	endpoint, err := container.PortEndpoint(ctx, "3306/tcp", "")
	if err != nil {
		t.Fatalf("failed to get container endpoint: %v", err)
	}
	host, portString, err := net.SplitHostPort(endpoint)
	if err != nil {
		t.Fatalf("failed to parse container endpoint %q: %v", endpoint, err)
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		t.Fatalf("failed to parse container port %q: %v", portString, err)
	}

	persistence, err := NewMariadbPersistence(config.DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     "devicestore",
		Password: "devicestore",
		Database: "devicestore",
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	t.Cleanup(func() {
		_ = persistence.db.Close()
	})

	applyMigrations(t, persistence.db)

	return persistence
}

// applyMigrations runs every migrations/VNNN.sql file, in order, against db.
func applyMigrations(t *testing.T, db *sql.DB) {
	t.Helper()
	migrationsDir := filepath.Join("..", "..", "..", "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("failed to read migrations directory: %v", err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)

	for _, file := range files {
		content, err := os.ReadFile(filepath.Join(migrationsDir, file))
		if err != nil {
			t.Fatalf("failed to read migration %s: %v", file, err)
		}
		for _, stmt := range splitSQLStatements(string(content)) {
			if _, err := db.Exec(stmt); err != nil {
				t.Fatalf("failed to apply migration %s, statement %q: %v", file, stmt, err)
			}
		}
	}
}

// splitSQLStatements splits a SQL script into individual statements, honoring
// MySQL's "DELIMITER" client directive (used by V016.sql to define a stored
// function whose body contains semicolons of its own).
func splitSQLStatements(script string) []string {
	delimiter := ";"
	var statements []string
	var current strings.Builder

	for _, line := range strings.Split(script, "\n") {
		trimmed := strings.TrimSpace(line)
		if upper := strings.ToUpper(trimmed); strings.HasPrefix(upper, "DELIMITER ") {
			delimiter = strings.TrimSpace(trimmed[len("DELIMITER "):])
			continue
		}
		current.WriteString(line)
		current.WriteString("\n")
		if strings.HasSuffix(trimmed, delimiter) {
			stmt := strings.TrimSpace(current.String())
			stmt = strings.TrimSuffix(stmt, delimiter)
			stmt = strings.TrimSpace(stmt)
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		}
	}
	return statements
}
