// Copyright (c) 2023 BVK Chaitanya

package kvtests

import (
	"context"
	"fmt"
	"log"

	"github.com/bvkgo/kv"
	"github.com/bvkgo/kv/kvtests/internal"
)

func RunTemplate(ctx context.Context, template string, db ...kv.Database) error {
	if _, err := internal.RunTemplateTest(ctx, template, db...); err != nil {
		return err
	}
	return nil
}

func RunBasicOps(ctx context.Context, db kv.Database) error {
	for k, v := range BasicOpsTemplateMap {
		log.Printf("*** %s ***", k)
		if err := RunTemplate(ctx, v, db); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
		if err := RunTemplate(ctx, v, db); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
		if err := RunTemplate(ctx, v, db); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
		if err := RunTemplate(ctx, v, db); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
		if err := RunTemplate(ctx, v, db); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
		if err := RunTemplate(ctx, v, db); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
		if err := RunTemplate(ctx, v, db); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
		if err := RunTemplate(ctx, v, db); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
		if err := RunTemplate(ctx, v, db); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
	}
	return nil
}

func RunTxOps(ctx context.Context, db kv.Database) error {
	Clear(ctx, db)
	for k, v := range TxOpsTemplateMap {
		if err := RunTemplate(ctx, v, db); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
		Clear(ctx, db)
		if err := RunTemplate(ctx, v, db); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
		Clear(ctx, db)
		if err := RunTemplate(ctx, v, db); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
		Clear(ctx, db)
		if err := RunTemplate(ctx, v, db); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
		Clear(ctx, db)
		if err := RunTemplate(ctx, v, db); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
		Clear(ctx, db)
		if err := RunTemplate(ctx, v, db); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
		Clear(ctx, db)
		if err := RunTemplate(ctx, v, db); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
		Clear(ctx, db)
		if err := RunTemplate(ctx, v, db); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
		Clear(ctx, db)
		if err := RunTemplate(ctx, v, db); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
		Clear(ctx, db)
		if err := RunTemplate(ctx, v, db); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
	}
	return nil
}
