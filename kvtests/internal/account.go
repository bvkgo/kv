// Copyright (c) 2023 BVK Chaitanya

package internal

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/bvkgo/kv"
)

type Account struct {
	ID      int
	Balance int64
}

func LoadAccount(key string, value io.Reader) (*Account, error) {
	a := new(Account)
	if _, err := fmt.Sscanf(key, "/accounts/%d", &a.ID); err != nil {
		return nil, err
	}
	if _, err := fmt.Fscanf(value, "%d", &a.Balance); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *Account) Reload(ctx context.Context, r kv.Getter) error {
	v, err := r.Get(ctx, a.Key())
	if err != nil {
		return err
	}
	var balance int64
	if _, err := fmt.Fscanf(v, "%d", &balance); err != nil {
		return err
	}
	a.Balance = balance
	return nil
}

func (a *Account) Key() string {
	return fmt.Sprintf("/accounts/%06d", a.ID)
}

func (a *Account) Value() io.Reader {
	return strings.NewReader(fmt.Sprintf("%d", a.Balance))
}

func (a *Account) Save(ctx context.Context, w kv.Setter) error {
	return w.Set(ctx, a.Key(), a.Value())
}
