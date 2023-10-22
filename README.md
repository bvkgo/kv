# An uniform API for many different Key-Value Databases

[![PkgGoDev](https://pkg.go.dev/badge/bvkgo/kv)](https://pkg.go.dev/github.com/bvkgo/kv)

This module defines an uniform API for key-value databases with snapshots and
transactions. API is aimed to be simple, unambiguous and well-defined. Multiple
database backends should be able to define adapters to support this API.

This package contains only the API interfaces and common utility
functions. Real key-value databases implementing this API are defined in
separate packages. For example, see
[github.com/bvkgo/kv/kvmemdb](https://pkg.go.dev/github.com/bvkgo/kv/kvmemdb).

## NOTES

### Empty string `""` is not a valid key

This API *requires* that empty string `""` cannot be a valid
key. [Iterator](https://pkg.go.dev/github.com/bvkgo/kv#Iterator) API uses empty
strings to indicate the end of key range.
