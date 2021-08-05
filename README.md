# An API for Key-Value Databases

[![PkgGoDev](https://pkg.go.dev/badge/bvkgo/kv)](https://pkg.go.dev/github.com/bvkgo/kv)

This module attempts to define a general-purpose api for key-value databases
with transactions. The API should be simple, useful, unambiguous and
well-defined. Multiple database backends should be able to implement the API
for interchangeability.

Only API interfaces are defined in this module. Real key-value databases
implementing this API are defined outside as separate modules. For example, see
[github.com/bvkgo/kvmemdb](https://pkg.go.dev/github.com/bvkgo/kvmemdb).

## Use `string`s instead of `[]byte` slices

Many key-value packages use `[]byte` for keys or values or for both, but this
module uses `string` data type for both keys and values.

`[]byte` slices are writable, so, when multiple components reference a `[]byte`
slice, they are sharing writable memory.  If `[]byte` slices are used as keys
or values, database implementation must make a private copy of the keys or
values. When not copied, users could modify the `[]byte` slice and break the
internal invariants of the database.

Using `string` type avoids the above issues. Since `string` values are
read-only `[]byte` slices, they can be referenced by multiple components and
goroutines simultaneously.  They can also hold any arbitrary bytes and are not
restricted to just utf8 characters.

## Empty string `""` is not a valid key

Empty strings `""` are treated as *invalid* keys. All implementations are
required to disallow reads and writes to empty keys.

There are not many usecases where users need to use empty string as a key. An
unique uuid-literal can serve the same purpose in most of the cases.

[Iterator](https://pkg.go.dev/github.com/bvkgo/kv#Iterator) API uses empty
strings to indicate the largest key (when ascending) or the smallest key (when
descending) or all of the keys in the database.
