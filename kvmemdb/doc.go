// Copyright (c) 2023 BVK Chaitanya

/*
Package kvmem implements an in-memory key-value database with snapshots and
read/write transactions. Snapshots can be taken periodically to backup and
restore the database.

Database can be used by multiple goroutines simultaneously, but the Snapshots
and Transactions are *not* thread-safe and can only be used by a single
goroutine.

		DESIGN NOTES

	 	All key-value data is stored in a shared sync.Map instance.
*/
package kvmemdb
