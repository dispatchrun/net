// Package postgres provides an example of a Postgres client compiled to
// GOOS=wasip1.
//
// The test demonstrates how to configure a custom dialer on the postgres client
// to use the dial functions implemented in github.com/stealthrocket/net/wasip1.
//
// When compiling to other targets than GOOS=wasip1, importing this package has
// no effect.
package postgres
