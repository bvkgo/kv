# An API for Key-Value Databases

[![PkgGoDev](https://pkg.go.dev/badge/bvkgo/kv)](https://pkg.go.dev/github.com/bvkgo/kv)

This module attempts to define a general-purpose api for key-value databases
with transactions. The API should be simple, useful, unambiguous and
well-defined. Multiple database backends must be able to implement the API for
interchangeability.

Only API interfaces are defined in this module. Real key-value databases
implementing this API are defined outside as separate modules. Unit tests for
all database implementations may be added in future.

## On `string` vs `[]byte`

Many key-value packages use `[]byte` for keys or values or for both. This
module uses `string` data type for both keys and values.

In Go, `[]byte` slices are writable, so when multiple components reference a
`[]byte` slice, they are sharing writable memory.  If `[]byte` slices are used
as keys or values, database implementation must make a private copy of the keys
or values. When not copied, users could modify the `[]byte` slice and break
internal invariants of the database.

Using `string` type avoids above issues. Since `string`s are read-only `[]byte`
slices in Go, they can be referenced by multiple components and goroutines
simultaneously.  They can also hold any arbitrary bytes and are not restricted
to just utf8 characters.

## Empty string `""` as a key

One potential downside to using `string` for keys is, users could use empty
string `""` as a key. This has a few implications with range selection, in the
[Iterator](https://pkg.go.dev/github.com/bvkgo/kv#Iterator) API, but they are
minor.

Whether an empty string is allowed as a valid key or not is left to the
database implementations. For example,
[github.com/bvkgo/kvmemdb](https://pkg.go.dev/github.com/bvkgo/kvmemdb) allows for both
modes.
