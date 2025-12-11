package test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

const (
	testAdminDSN = "postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	testDBName   = "loyalty_integration_test"
)

func setupTestDatabase(t *testing.T) string {
	ctx := context.Background()

	adminConfig, err := pgx.ParseConfig(testAdminDSN)
	if err != nil {
		t.Fatalf("Failed to parse admin DSN: %v", err)
	}

	adminDB := stdlib.OpenDB(*adminConfig)
	defer adminDB.Close()

	if err := adminDB.PingContext(ctx); err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
	}

	var dbExists int
	err = adminDB.QueryRowContext(ctx,
		"SELECT 1 FROM pg_database WHERE datname = $1", testDBName).Scan(&dbExists)

	if err == sql.ErrNoRows {
		_, err = adminDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", testDBName))
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
	}

	testDSN := os.Getenv("TEST_DATABASE_URI")
	if testDSN == "" {
		testDSN = fmt.Sprintf("postgresql://postgres:postgres@localhost:5432/%s?sslmode=disable", testDBName)
	}

	testConfig, err := pgx.ParseConfig(testDSN)
	if err != nil {
		t.Fatalf("Failed to parse test DSN: %v", err)
	}

	testDB := stdlib.OpenDB(*testConfig)
	defer testDB.Close()

	if err := testDB.PingContext(ctx); err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	_, err = testDB.ExecContext(ctx, "TRUNCATE TABLE gophermart.users, gophermart.orders, gophermart.balance_transactions CASCADE")
	if err != nil {
		t.Logf("Warning: failed to truncate tables (may not exist yet): %v", err)
	}

	return testDSN
}

type testEnvironment struct {
	gophermartURL  string
	gophermartPort int
	gophermartProc *process
	accrualMock    *AccrualMock
	accrualProc    *process
	dbURI          string
	binaryPath     string
	httpClient     *http.Client
}

func setupTestEnvironment(t *testing.T) *testEnvironment {
	ctx := context.Background()

	dbURI := setupTestDatabase(t)

	binaryPath, err := buildBinary()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	accrualMock, accrualProc, err := startAccrualMock(ctx)
	if err != nil {
		t.Fatalf("Failed to start accrual mock: %v", err)
	}

	gophermartProc, gophermartPort, err := startGophermart(ctx, binaryPath, dbURI, accrualMock.URL())
	if err != nil {
		accrualProc.Stop()
		t.Fatalf("Failed to start gophermart: %v", err)
	}

	return &testEnvironment{
		gophermartURL:  fmt.Sprintf("http://localhost:%d", gophermartPort),
		gophermartPort: gophermartPort,
		gophermartProc: gophermartProc,
		accrualMock:    accrualMock,
		accrualProc:    accrualProc,
		dbURI:          dbURI,
		binaryPath:     binaryPath,
		httpClient:     makeHTTPClient(),
	}
}

func (env *testEnvironment) teardown() {
	if env.gophermartProc != nil {
		_ = env.gophermartProc.Stop()
	}
	if env.accrualProc != nil {
		_ = env.accrualProc.Stop()
	}
}
