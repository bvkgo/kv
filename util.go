// Copyright (c) 2023 BVK Chaitanya

package kv

import (
	"context"
	"io"
)

// Close is a helper function that invokes Close method on the input Iterator.
//
// Some key-value store implementations may need to release resources, so this
// generalized helper function can be used without access to concrete data type
// of the underlying object.
func Close(v Iterator) error {
	if c, ok := v.(io.Closer); ok {
		c.Close()
		return nil
	}

	if c, ok := v.(interface{ Close() }); ok {
		c.Close()
		return nil
	}
	return nil
}

func WithSnapshot(ctx context.Context, db Database, f func(context.Context, Snapshot) error) error {
	snap, err := db.NewSnapshot(ctx)
	if err != nil {
		return err
	}
	defer snap.Discard(ctx)

	if err := f(ctx, snap); err != nil {
		return err
	}
	return nil
}

func WithTransaction(ctx context.Context, db Database, f func(context.Context, Transaction) error) error {
	tx, err := db.NewTransaction(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := f(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}
