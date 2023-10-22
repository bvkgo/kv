// Copyright (c) 2023 BVK Chaitanya

/*
Package kvmemdb implements an in-memory key-value database with snapshots and
read/write transactions. Snapshots can be taken periodically to backup and
restore the database.

Database can be used by multiple goroutines simultaneously, however, individual
Snapshot and Transaction objects are not thread-safe; they can only be used by
a single goroutine.
*/
package kvmemdb
