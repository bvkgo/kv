// Copyright (c) 2023 BVK Chaitanya

package kvtests

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/bvkgo/kv"
	"github.com/bvkgo/kv/kvtests/internal"
	"golang.org/x/sync/errgroup"
)

// BankTest uses multiple goroutines to perform transactions simultaneously
// while verifying that expected invariant holds true at every snapshot. This
// test runs till input context is canceled or verification check has failed.
//
// Multiple goroutines randomly transfer amounts between the accounts in
// parallel. The goroutine that is running the tests will take repeated
// snapshots of the database and verifies that sum of balances across all
// accounts doesn't change over time.
type BankTest struct {
	DB kv.Database

	// InitializeDB when true clears the database and initializes it with random
	// accounts.
	InitializeDB bool

	minBalance   int64
	numAccounts  int
	totalBalance int64
}

func (b *BankTest) setDefaults() {
	if b.minBalance == 0 {
		b.minBalance = 10
	}
	if b.numAccounts == 0 {
		b.numAccounts = 1000
	}
}

func (b *BankTest) TotalBalance() int64 {
	return b.totalBalance
}

func (b *BankTest) initializeDB(ctx context.Context) error {
	if err := Clear(ctx, b.DB); err != nil {
		return fmt.Errorf("could not clear the database: %w", err)
	}

	// Initialize the database.
	b.totalBalance = int64(0)
	initDB := func(ctx context.Context, tx kv.Transaction) error {
		for i := 0; i < b.numAccounts; i++ {
			a := &internal.Account{
				ID:      i,
				Balance: b.minBalance + int64(rand.Int31n(math.MaxInt32)),
			}
			if err := a.Save(ctx, tx); err != nil {
				return err
			}
			b.totalBalance += a.Balance
		}
		return nil
	}
	if err := b.DB.WithTransaction(ctx, initDB); err != nil {
		return err
	}
	return nil
}

func (b *BankTest) FindTotalBalance(ctx context.Context) (int64, error) {
	var totalBalance int64
	totalDB := func(ctx context.Context, s kv.Snapshot) error {
		it, err := s.Scan(ctx)
		if err != nil {
			return err
		}
		defer kv.Close(it)

		for k, v, ok := it.Current(ctx); ok; k, v, ok = it.Next(ctx) {
			a, err := internal.LoadAccount(k, v)
			if err != nil {
				return err
			}
			totalBalance += a.Balance
		}

		if err := it.Err(); err != nil {
			return err
		}
		return nil
	}
	if err := b.DB.WithSnapshot(ctx, totalDB); err != nil {
		return 0, err
	}
	return totalBalance, nil
}

// Run executes the test with nclient goroutines performing database
// transactions simultaneously.
func (b *BankTest) Run(ctx context.Context, nclients int) error {
	b.setDefaults()

	if b.InitializeDB {
		if err := b.initializeDB(ctx); err != nil {
			return err
		}
	} else if total, err := b.FindTotalBalance(ctx); err != nil {
		return err
	} else {
		b.totalBalance = total
	}

	eg, egCtx := errgroup.WithContext(ctx)
	for i := 0; i < nclients; i++ {
		eg.Go(func() error {
			for egCtx.Err() == nil {
				if err := b.updateDB(egCtx); err != nil {
					// log.Printf("warning: update failed: %v", err)
				}
			}
			return nil
		})
	}

	for i := 1; egCtx.Err() == nil; i++ {
		if err := b.verifyDB(egCtx); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (b *BankTest) updateDB(ctx context.Context) error {
	updateDB := func(ctx context.Context, tx kv.Transaction) error {
		src := &internal.Account{ID: rand.Intn(b.numAccounts)}
		if err := src.Reload(ctx, tx); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}

		dst := &internal.Account{ID: rand.Intn(b.numAccounts)}
		if err := dst.Reload(ctx, tx); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
			// We will create a new account for dst.
		}

		if src.ID == dst.ID {
			return nil
		}
		if src.Balance < 2*b.minBalance {
			return nil
		}

		// Every now and then, we Delete the src account by completely transferring
		// it's balance to the dst account.
		//
		// If src account is not being deleted then we want to ensure both src and
		// dst accounts to have minBalance

		amount := src.Balance
		if v := rand.Int(); v%5 != 0 {
			amount = b.minBalance + rand.Int63n(src.Balance-2*b.minBalance+1) - 1
		}

		src.Balance -= amount
		dst.Balance += amount

		if src.Balance == 0 {
			if err := tx.Delete(ctx, src.Key()); err != nil {
				return err
			}
		} else {
			if err := src.Save(ctx, tx); err != nil {
				return err
			}
		}

		if err := dst.Save(ctx, tx); err != nil {
			return err
		}

		// log.Printf("%v: transferring %d from %s to %s", tx, amount, src.Key(), dst.Key())
		return nil
	}
	return b.DB.WithTransaction(ctx, updateDB)
}

func (b *BankTest) verifyDB(ctx context.Context) error {
	verifyDB := func(ctx context.Context, s kv.Snapshot) error {
		it, err := s.Scan(ctx)
		if err != nil {
			return err
		}
		defer kv.Close(it)

		count := 0
		total := int64(0)
		for k, v, ok := it.Current(ctx); ok; k, v, ok = it.Next(ctx) {
			a, err := internal.LoadAccount(k, v)
			if err != nil {
				return err
			}
			total += a.Balance
			count++
		}

		if err := it.Err(); err != nil {
			return err
		}

		if total != b.totalBalance {
			missing := b.totalBalance - total
			return fmt.Errorf("unexpected total balance (missing %d): want %d, got %d", missing, b.totalBalance, total)
		}
		log.Printf("snapshot has %d balance in %d accounts", b.totalBalance, count)
		return nil
	}
	return b.DB.WithSnapshot(ctx, verifyDB)
}