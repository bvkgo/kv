// Copyright (c) 2023 BVK Chaitanya

package kvmemdb

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/bvkgo/kv/kvtests"
)

func TestBankTransactions(t *testing.T) {
	ctx := context.Background()
	db := New()

	duration := 60 * time.Second

	tctx, tcancel := context.WithTimeout(ctx, duration)
	defer tcancel()

	// if limit, ok := t.Deadline(); ok {
	// 	nctx, tcancel := context.WithDeadline(ctx, limit.Add(-time.Second))
	// 	defer tcancel()
	// 	tctx = nctx
	// }

	b := kvtests.BankTest{
		DB:           db,
		InitializeDB: true,
	}

	if err := b.Run(tctx, 1000); err != nil {
		t.Fatal(err)
	}
	balance := b.TotalBalance()

	t.Logf("compaction deleted %d items", Compact(ctx, db))

	file := filepath.Join(t.TempDir(), "backup.db")
	if err := Backup(db, file); err != nil {
		t.Fatal(err)
	}

	{
		db, err := Restore(file)
		if err != nil {
			t.Fatal(err)
		}

		tctx, tcancel := context.WithTimeout(ctx, duration)
		defer tcancel()

		b := kvtests.BankTest{
			DB:           db,
			InitializeDB: false,
		}

		if v, err := b.FindTotalBalance(ctx); err != nil {
			t.Fatal(err)
		} else if v != balance {
			t.Fatalf("unexpected balance; want %d, got %d", balance, v)
		}

		if err := b.Run(tctx, 100); err != nil {
			t.Fatal(err)
		}
	}
}
