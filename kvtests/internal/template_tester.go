// Copyright (c) 2023 BVK Chaitanya

package internal

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/bvkgo/kv"
)

type TemplateTest struct {
	Steps []*TemplateStep
}

func ParseTemplateTest(r io.Reader) (*TemplateTest, error) {
	var steps []*TemplateStep

	s := bufio.NewScanner(r)
	s.Split(bufio.SplitFunc(bufio.ScanLines))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		step, err := ParseTemplateStep(line)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	t := &TemplateTest{
		Steps: steps,
	}
	return t, nil
}

// WithKeyPrefix returns a new test with the given prefix added to all
// keys. Templates with `scan` operator or empty values for `begin` and `end`
// are not supported.
func (t *TemplateTest) WithKeyPrefix(prefix string) (*TemplateTest, error) {
	steps := make([]*TemplateStep, 0, len(t.Steps))
	for _, s := range t.Steps {
		if s.Op == `scan` {
			return nil, fmt.Errorf("templates with `scan` operator are not supported")
		}
		if s.Op == `ascend` || s.Op == `descend` {
			if s.Begin == "" || s.End == "" {
				return nil, fmt.Errorf("templates with empty `begin` or `end` values are not supported")
			}
		}

		ns := s
		if len(ns.Key) != 0 {
			ns.Key = prefix + s.Key
		}
		if len(ns.Begin) != 0 {
			ns.Begin = prefix + s.Begin
		}
		if len(ns.End) != 0 {
			ns.End = prefix + s.End
		}
	}
	nt := &TemplateTest{
		Steps: steps,
	}
	return nt, nil
}

// NumDatabase returns number of transactions in the template.
func (t *TemplateTest) NumDatabase() int {
	m := make(map[string]struct{})
	for _, s := range t.Steps {
		m[s.Database] = struct{}{}
	}
	return len(m)
}

func (t *TemplateTest) Databases() []string {
	var dbs []string

	m := make(map[string]struct{})
	for _, s := range t.Steps {
		if len(s.Database) == 0 {
			continue
		}
		if _, ok := m[s.Database]; ok {
			continue
		}
		m[s.Database] = struct{}{}
		dbs = append(dbs, s.Database)
	}
	return dbs
}

// RunTemplateTest runs all database operations in the input template serially
// one-after-another in the order defined in the template.
func RunTemplateTest(ctx context.Context, text string, dbs ...kv.Database) ([]*TemplateStepResult, error) {
	test, err := ParseTemplateTest(strings.NewReader(text))
	if err != nil {
		return nil, err
	}

	dbMap := make(map[string]kv.Database)
	itMap := make(map[string]kv.Iterator)
	txMap := make(map[string]kv.Transaction)
	snapMap := make(map[string]kv.Snapshot)
	defer func() {
		for _, tx := range txMap {
			tx.Rollback(ctx)
		}
		for _, snap := range snapMap {
			snap.Discard(ctx)
		}
		for _, it := range itMap {
			kv.Close(it)
		}
	}()

	// Configure databases.
	dbkeys := test.Databases()
	if len(dbs) != len(dbkeys) {
		return nil, fmt.Errorf("template needs %d Databases %v", len(dbkeys), dbkeys)
	}
	for i, db := range dbs {
		dbMap[dbkeys[i]] = db
	}

	var results []*TemplateStepResult
	for _, s := range test.Steps {
		var db kv.Database
		var tx kv.Transaction
		var snap kv.Snapshot
		var iter kv.Iterator
		if len(s.Database) != 0 {
			db, _ = dbMap[s.Database]
		}
		if len(s.Transaction) != 0 {
			tx, _ = txMap[s.Transaction]
		}
		if len(s.Snapshot) != 0 {
			snap, _ = snapMap[s.Snapshot]
		}
		if len(s.Iterator) != 0 {
			iter, _ = itMap[s.Iterator]
		}

		result, err := s.Run(ctx, db, tx, snap, iter)
		if err != nil {
			return nil, err
		}

		if result.Transaction != nil {
			txMap[s.Transaction] = result.Transaction
		}
		if result.Iterator != nil {
			itMap[s.Iterator] = result.Iterator
		}
		if result.Snapshot != nil {
			snapMap[s.Snapshot] = result.Snapshot
		}
		results = append(results, result)

		if s.Op == `rollback` || s.Op == `commit` {
			delete(txMap, s.Transaction)
		}
		if s.Op == `discard` {
			delete(snapMap, s.Snapshot)
		}
	}

	var txKeys []string
	for k := range txMap {
		txKeys = append(txKeys, k)
	}
	if len(txKeys) != 0 {
		log.Printf("lingering transactions %v will be rollback-ed automatically", txKeys)
	}
	return results, nil
}
