// Copyright (c) 2023 BVK Chaitanya

package internal

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/bvkgo/kv"
)

type TemplateRuntime struct {
	dbMap       map[string]kv.Database
	txMap       map[string]kv.Transaction
	snapMap     map[string]kv.Snapshot
	itMap       map[string]kv.Iterator
	txIterMap   map[string][]string
	snapIterMap map[string][]string
}

func newTemplateRuntime(steps []*TemplateStep, dbs ...kv.Database) (*TemplateRuntime, error) {
	var dbkeys []string
	m := make(map[string]struct{})
	for _, s := range steps {
		if len(s.Database) == 0 {
			continue
		}
		if _, ok := m[s.Database]; ok {
			continue
		}
		m[s.Database] = struct{}{}
		dbkeys = append(dbkeys, s.Database)
	}

	// Configure databases.
	if len(dbs) != len(dbkeys) {
		return nil, fmt.Errorf("template needs %d Databases %v", len(dbkeys), dbkeys)
	}

	rt := &TemplateRuntime{
		dbMap:       make(map[string]kv.Database),
		txMap:       make(map[string]kv.Transaction),
		itMap:       make(map[string]kv.Iterator),
		snapMap:     make(map[string]kv.Snapshot),
		txIterMap:   make(map[string][]string),
		snapIterMap: make(map[string][]string),
	}

	for i, db := range dbs {
		rt.dbMap[dbkeys[i]] = db
	}
	return rt, nil
}

func (rt *TemplateRuntime) Close() {
	for _, it := range rt.itMap {
		kv.Close(it)
	}
	for k, tx := range rt.txMap {
		tx.Rollback(context.Background())
		log.Printf("lingering transaction %q is rollbacked automatically", k)
	}
	for k, snap := range rt.snapMap {
		snap.Discard(context.Background())
		log.Printf("lingering snapshot %q is discarded automatically", k)
	}
}

// RunStep executes the database command on a user DB.
func (rt *TemplateRuntime) RunStep(ctx context.Context, s *TemplateStep) (*TemplateStepResult, error) {
	result := &TemplateStepResult{
		Step: s,
	}

	// log.Printf("run-step: %s", s)

	switch s.Op {
	default:
		return nil, fmt.Errorf("invalid/unrecognized op %q", s.Op)

	case "get":
		if len(s.Transaction) != 0 {
			tx := rt.txMap[s.Transaction]
			result.Value, result.Status = tx.Get(ctx, s.Key)
		}
		if len(s.Snapshot) != 0 {
			snap := rt.snapMap[s.Snapshot]
			result.Value, result.Status = snap.Get(ctx, s.Key)
		}

	case "set":
		tx := rt.txMap[s.Transaction]
		result.Status = tx.Set(ctx, s.Key, strings.NewReader(s.Value))
	case "delete":
		tx := rt.txMap[s.Transaction]
		result.Status = tx.Delete(ctx, s.Key)

	case "ascend":
		if len(s.Transaction) != 0 {
			tx := rt.txMap[s.Transaction]
			result.Iterator, result.Status = tx.Ascend(ctx, s.Begin, s.End)
			rt.itMap[s.Iterator] = result.Iterator
			rt.txIterMap[s.Transaction] = append(rt.txIterMap[s.Transaction], s.Iterator)
			break
		}
		if len(s.Snapshot) != 0 {
			snap := rt.snapMap[s.Snapshot]
			result.Iterator, result.Status = snap.Ascend(ctx, s.Begin, s.End)
			rt.itMap[s.Iterator] = result.Iterator
			rt.snapIterMap[s.Snapshot] = append(rt.snapIterMap[s.Snapshot], s.Iterator)
			break
		}
		return nil, fmt.Errorf("%q has no tx or snap name", s.Op)

	case "descend":
		if len(s.Transaction) != 0 {
			tx := rt.txMap[s.Transaction]
			result.Iterator, result.Status = tx.Descend(ctx, s.Begin, s.End)
			rt.itMap[s.Iterator] = result.Iterator
			rt.txIterMap[s.Transaction] = append(rt.txIterMap[s.Transaction], s.Iterator)
			break
		}
		if len(s.Snapshot) != 0 {
			snap := rt.snapMap[s.Snapshot]
			result.Iterator, result.Status = snap.Descend(ctx, s.Begin, s.End)
			rt.itMap[s.Iterator] = result.Iterator
			rt.snapIterMap[s.Snapshot] = append(rt.snapIterMap[s.Snapshot], s.Iterator)
			break
		}
		return nil, fmt.Errorf("%q has no tx or snap name", s.Op)

	case "scan":
		if len(s.Transaction) != 0 {
			tx := rt.txMap[s.Transaction]
			result.Iterator, result.Status = tx.Scan(ctx)
			rt.itMap[s.Iterator] = result.Iterator
			rt.txIterMap[s.Transaction] = append(rt.txIterMap[s.Transaction], s.Iterator)
			break
		}
		if len(s.Snapshot) != 0 {
			snap := rt.snapMap[s.Snapshot]
			result.Iterator, result.Status = snap.Scan(ctx)
			rt.itMap[s.Iterator] = result.Iterator
			rt.snapIterMap[s.Snapshot] = append(rt.snapIterMap[s.Snapshot], s.Iterator)
			break
		}
		return nil, fmt.Errorf("%q has no tx or snap name", s.Op)

	case "current":
		iter := rt.itMap[s.Iterator]
		if k, v, ok := iter.Current(ctx); ok {
			result.Key, result.Value, result.Status = k, v, nil
		} else {
			result.Key, result.Value, result.Status = k, v, iter.Err()
		}

	case "next":
		iter := rt.itMap[s.Iterator]
		if k, v, ok := iter.Next(ctx); ok {
			result.Key, result.Value, result.Status = k, v, nil
		} else {
			result.Key, result.Value, result.Status = k, v, iter.Err()
		}

	case "new-transaction":
		db := rt.dbMap[s.Database]
		result.Transaction, result.Status = db.NewTransaction(ctx)
		rt.txMap[s.Transaction] = result.Transaction

	case "commit":
		for _, it := range rt.txIterMap[s.Transaction] {
			kv.Close(rt.itMap[it])
			delete(rt.itMap, it)
		}
		tx := rt.txMap[s.Transaction]
		result.Status = tx.Commit(ctx)

	case "rollback":
		for _, it := range rt.txIterMap[s.Transaction] {
			kv.Close(rt.itMap[it])
			delete(rt.itMap, it)
		}
		tx := rt.txMap[s.Transaction]
		result.Status = tx.Rollback(ctx)

	case "new-snapshot":
		db := rt.dbMap[s.Database]
		result.Snapshot, result.Status = db.NewSnapshot(ctx)
		rt.snapMap[s.Snapshot] = result.Snapshot

	case "discard":
		for _, it := range rt.snapIterMap[s.Snapshot] {
			kv.Close(rt.itMap[it])
			delete(rt.itMap, it)
		}
		snap := rt.snapMap[s.Snapshot]
		result.Status = snap.Discard(ctx)
	}

	if err := s.checkStatus(result); err != nil {
		return nil, err
	}
	if err := s.checkKey(result); err != nil {
		return nil, err
	}
	if err := s.checkValue(result); err != nil {
		return nil, err
	}

	return result, nil
}
