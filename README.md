# An API for Key-Value Databases

[![PkgGoDev](https://pkg.go.dev/badge/bvkgo/kv)](https://pkg.go.dev/github.com/bvkgo/kv)

This package attempts to define a clean and minimal api for key-value databases
with transactions. Only the API interfaces are defined in this package.

Real key-value databases implementing this API will be defined outside, as
separate, independent packages so that there are no external dependencies. Unit
tests for all database implementations may be added in future.

Most of the effort is spent on making the API simple, useful, unambiguous and
well-defined.
