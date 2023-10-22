// Copyright (c) 2023 BVK Chaitanya

package kvmemdb

import (
	"context"
	"testing"
	"time"

	"github.com/bvkgo/kv/kvtests"
)

func TestBasicTest(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := New()
	if err := kvtests.RunTemplate(ctx, kvtests.TxGetSetDelete, db); err != nil {
		t.Fatal("TxGetSetDelete", err)
	}
	if err := kvtests.RunTemplate(ctx, kvtests.TxAscendDescendScanEmpty, db); err != nil {
		t.Fatal("TxAscendDescendScanEmpty", err)
	}
	if err := kvtests.RunTemplate(ctx, kvtests.TxAscendDescendInvalid, db); err != nil {
		t.Fatal("TxAscendDescendInvalid", err)
	}
	if err := kvtests.RunTemplate(ctx, kvtests.TxAscendEmpty, db); err != nil {
		t.Fatal("TxAscendEmpty", err)
	}
	if err := kvtests.RunTemplate(ctx, kvtests.TxDescendEmpty, db); err != nil {
		t.Fatal("TxDescendEmpty", err)
	}
	if err := kvtests.RunTemplate(ctx, kvtests.TxAscendNonEmptyRange, db); err != nil {
		t.Fatal("TxAscendNonEmptyRange", err)
	}
	if err := kvtests.RunTemplate(ctx, kvtests.TxDescendNonEmptyRange, db); err != nil {
		t.Fatal("TxDescendNonEmptyRange", err)
	}
	if err := kvtests.RunTemplate(ctx, kvtests.TxAscendOneEmptyRange, db); err != nil {
		t.Fatal("TxAscendOneEmptyRange", err)
	}
	if err := kvtests.RunTemplate(ctx, kvtests.TxDescendOneEmptyRange, db); err != nil {
		t.Fatal("TxDescendOneEmptyRange", err)
	}
}

func TestTxSemantics(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := New()
	if err := kvtests.RunTemplate(ctx, kvtests.SerializedTxCommits, db); err != nil {
		t.Fatal("SerializedTxCommits", err)
	}
	kvtests.Clear(ctx, db)
	if err := kvtests.RunTemplate(ctx, kvtests.SerializedTxCommitsAndRollbacks, db); err != nil {
		t.Fatal("SerializedTxCommitsAndRollbacks", err)
	}
	kvtests.Clear(ctx, db)
	if err := kvtests.RunTemplate(ctx, kvtests.NonConflictingTxes, db); err != nil {
		t.Fatal("NonConflictingTxes", err)
	}
	kvtests.Clear(ctx, db)
	if err := kvtests.RunTemplate(ctx, kvtests.ConflictingReadOnlyTxes, db); err != nil {
		t.Fatal("ConflictingReadOnlyTxes", err)
	}
	kvtests.Clear(ctx, db)
	if err := kvtests.RunTemplate(ctx, kvtests.ConflictingReadWriteTxes, db); err != nil {
		t.Fatal("ConflictingReadWriteTxes", err)
	}
	kvtests.Clear(ctx, db)
	if err := kvtests.RunTemplate(ctx, kvtests.ConflictingDeletes, db); err != nil {
		t.Fatal("ConflictingDeletes", err)
	}
	kvtests.Clear(ctx, db)
	if err := kvtests.RunTemplate(ctx, kvtests.NonConflictingDeletes, db); err != nil {
		t.Fatal("NonConflictingDeletes", err)
	}
	kvtests.Clear(ctx, db)
	if err := kvtests.RunTemplate(ctx, kvtests.AbortedReads, db); err != nil {
		t.Fatal("AbortedReads", err)
	}
	kvtests.Clear(ctx, db)
	if err := kvtests.RunTemplate(ctx, kvtests.RepeatedReads, db); err != nil {
		t.Fatal("RepeatedReads", err)
	}
}
