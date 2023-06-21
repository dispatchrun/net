//go:build wasip1

package postgres_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stealthrocket/net/wasip1"
)

func TestPostgres(t *testing.T) {
	config, err := pgx.ParseConfig("postgres://pqgotest:password@localhost:5432/test?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	// Here we configure a custom dialer for the postgres connector in order to
	// use the dial functions from github.com/stealthrocket/net/wasip1.
	config.DialFunc = wasip1.DialContext
	// Avoid using the default net.LookupHost function which is not currently
	// supported on GOOS=wasip1.
	config.LookupFunc = func(ctx context.Context, host string) (addrs []string, err error) {
		return []string{host}, nil
	}
	connect := stdlib.RegisterConnConfig(config)

	db, err := sql.Open("pgx", connect)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	row := db.QueryRow("select version()")
	ver := ""
	if err := row.Scan(&ver); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(ver, "PostgreSQL") {
		t.Errorf("wrong version returned by the mysql server: %q", ver)
	}
}
