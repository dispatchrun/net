// Package mysql configures the go-sql-driver/mysql package to support opening
// network connections when compiled to GOOS=wasip1.
//
// This package is intended to be imported as a nameless package with:
//
//	import (
//		_ "github.com/stealthrocket/net/mysql"
//	)
//
// The package is distributed as a separate module so only applications which
// use mysql need to take a dependency on this module.
//
// When compiling to other targets than GOOS=wasip1, importing this package has
// no effect.
package mysql
