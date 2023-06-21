// Package postgres provides an example of a Postgres client compiled to
// GOOS=wasip1.
//
// The test demonstrates how to configure a custom dial function on the postgres
// client to use the github.com/stealthrocket/net/wasip1 package.
//
// When compiling to other targets than GOOS=wasip1, importing this package has
// no effect.
package postgres
