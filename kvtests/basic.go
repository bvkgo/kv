// Copyright (c) 2023 BVK Chaitanya

package kvtests

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bvkgo/kv"
)

func BasicTest(ctx context.Context, db kv.Database) error {
	if err := Clear(ctx, db); err != nil {
		return err
	}

	basicChecks := func(ctx context.Context, tx kv.Transaction) error {
		if _, err := tx.Get(ctx, "/does/not/exist"); err == nil {
			return fmt.Errorf("get: wanted os.ErrNotExist, got nil")
		}
		if err := tx.Delete(ctx, "/does/not/exist"); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("get: wanted nil or os.ErrNotExist, got %v", err)
		}

		if err := tx.Set(ctx, "key1", strings.NewReader("value1")); err != nil {
			return fmt.Errorf("set: wanted nil, got %v", err)
		}
		if v, err := tx.Get(ctx, "key1"); err != nil {
			return fmt.Errorf("get: wanted nil, got %v", err)
		} else if data, err := io.ReadAll(v); err != nil {
			return fmt.Errorf("readall: wanted nil, got %v", err)
		} else if s := string(data); s != "value1" {
			return fmt.Errorf("get: wanted value1, got %q", s)
		}

		if err := tx.Delete(ctx, "key1"); err != nil {
			return fmt.Errorf("delete: wanted nil, got %v", err)
		}
		if _, err := tx.Get(ctx, "key1"); !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("get: wanted os.ErrNotExist, got %v", err)
		}

		return nil
	}
	if err := db.WithTransaction(ctx, basicChecks); err != nil {
		return err
	}

	return nil
}
