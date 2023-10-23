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
	if err := kvtests.RunBasicOps(ctx, db); err != nil {
		t.Fatal(err)
	}
}

func TestTxSemantics(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := New()
	if err := kvtests.RunTxOps(ctx, db); err != nil {
		t.Fatal(err)
	}
}
