// Copyright (c) 2023 BVK Chaitanya

package kvtests

import (
	"context"

	"github.com/bvkgo/kv"
)

// ClearDatabase deletes all key-value pairs in the database.
func Clear(ctx context.Context, db kv.Database) error {
	clear := func(ctx context.Context, tx kv.Transaction) error {
		it, err := tx.Scan(ctx)
		if err != nil {
			return err
		}
		defer kv.Close(it)

		for k, _, ok := it.Current(ctx); ok; k, _, ok = it.Next(ctx) {
			if err := tx.Delete(ctx, k); err != nil {
				return err
			}
		}

		if err := it.Err(); err != nil {
			return err
		}
		return nil
	}
	return kv.WithTransaction(ctx, db, clear)
}
