# FIXME: An API for Key-Value Databases

[![PkgGoDev](https://pkg.go.dev/badge/bvkgo/kv)](https://pkg.go.dev/github.com/bvkgo/kv)

This module attempts to define a general-purpose api for key-value databases
with transactions. The API should be simple, useful, unambiguous and
well-defined. Multiple database backends should be able to implement the API
for interchangeability.

Only API interfaces are defined in this module. Real key-value databases
implementing this API are defined outside as separate modules. For example, see
[github.com/bvkgo/kvmemdb](https://pkg.go.dev/github.com/bvkgo/kvmemdb).

## No `[]byte` slices

Many key-value packages use `[]byte` for keys and values, but this module uses
`string` and `io.Reader` types for keys and values respectively.

`[]byte` slices are writable, so, when multiple components reference a `[]byte`
slice, they are sharing writable memory.  If `[]byte` slices are used as keys
or values, database implementation must make a private copy. Otherwise, users
could modify the `[]byte` slice and break the internal invariants of the
database.

Using `string` type for keys avoids the above issues. Since `string`s are
immutable, they can be referenced by multiple components and goroutines
simultaneously.  They can also hold any arbitrary bytes and are not restricted
to just utf8 characters.

## Empty string `""` is not a valid key

Empty strings `""` cannot be used
keys. [Iterator](https://pkg.go.dev/github.com/bvkgo/kv#Iterator) API uses
empty strings to indicate all of the key range or the end of keys (i.e., the
largest key when ascending or the smallest key when descending).

There are not many usecases where users need to use an empty string as a
key. An unique uuid-literal can serve the same purpose in most of the cases.
