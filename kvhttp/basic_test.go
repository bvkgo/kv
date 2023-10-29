// Copyright (c) 2023 BVK Chaitanya

package kvhttp

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/bvkgo/kv/kvmemdb"
	"github.com/bvkgo/kv/kvtests"
)

func TestBasicTest(t *testing.T) {
	ctx := context.Background()

	s := httptest.NewServer(Handler(kvmemdb.New()))
	defer s.Close()

	addrURL, _ := url.Parse(s.URL)
	db := New(addrURL, s.Client())

	if err := kvtests.RunBasicOps(ctx, db); err != nil {
		t.Fatal(err)
	}
}

func TestTxSemantics(t *testing.T) {
	ctx := context.Background()

	s := httptest.NewServer(Handler(kvmemdb.New()))
	defer s.Close()

	addrURL, _ := url.Parse(s.URL)
	db := New(addrURL, s.Client())

	if err := kvtests.RunTxOps(ctx, db); err != nil {
		t.Fatal(err)
	}
}
