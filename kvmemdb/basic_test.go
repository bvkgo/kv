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
	if err := kvtests.BasicTest(ctx, db); err != nil {
		t.Fatal(err)
	}
}
