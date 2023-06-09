//go:build wasip1

package mysql_test

import (
	"database/sql"
	"testing"

	_ "github.com/stealthrocket/net/mysql"
)

func TestMySQL(t *testing.T) {
	db, err := sql.Open("mysql", "root:test@tcp(localhost:3306)/test")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	r, err := db.Query("select version()")
	if err != nil {
		t.Fatal(err)
	}
	_ = r
}
