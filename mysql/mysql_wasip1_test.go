//go:build wasip1

package mysql_test

import (
	"database/sql"
	"testing"

	// Importing this package configures go-sql-driver/mysql to use the dialer
	// from github.com/stealthrocket/net/wasip1.
	_ "github.com/stealthrocket/net/mysql"
)

func TestMySQL(t *testing.T) {
	db, err := sql.Open("mysql", "root:test@tcp(localhost:3306)/test")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	row := db.QueryRow("select version()")
	ver := ""
	if err := row.Scan(&ver); err != nil {
		t.Fatal(err)
	}
	if ver != "8.0.33" {
		t.Errorf("wrong version returned by the mysql server: %q", ver)
	}
}
