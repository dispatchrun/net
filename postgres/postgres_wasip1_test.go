//go:build wasip1

package postgres_test

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/lib/pq"
	"github.com/stealthrocket/net/wasip1"
)

func TestPostgres(t *testing.T) {
	connector, err := pq.NewConnector("user=pqgotest password=password dbname=test sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	deadline, _ := t.Deadline()
	// Here we configure a custom dialer for the postgres connector in order to
	// use the dial functions from github.com/stealthrocket/net/wasip1.
	connector.Dialer(&wasip1.Dialer{
		Deadline: deadline,
	})

	db := sql.OpenDB(connector)
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
