// Package memcache provides an example of a memcache client compiled to
// GOOS=wasip1.
//
// The test uses https://pkg.go.dev/github.com/bradfitz/gomemcache/memcache and
// interacts with a memcache server on localhost:11211. The docker-compose.yml
// file in the parent directory may be used to start a memcache server to run
// the test against.
//
// Note that at this time, the dependency is changed to https://github.com/mar4uk/gomemcache
// in order to get the changes from https://github.com/bradfitz/gomemcache/pull/158,
// which is required to customize the dial function.
//
// When compiling to other targets than GOOS=wasip1, importing this package has
// no effect.
package memcache
