// Copyright (c) 2023 BVK Chaitanya

package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/bvkgo/kv"
)

type TemplateOp string

var ops = []TemplateOp{
	"get",
	"set",
	"delete",

	"ascend",
	"descend",
	"scan",

	"current",
	"next",

	"new-transaction",
	"commit",
	"rollback",

	"new-snapshot",
	"discard",
}

// TemplateStep represents a sinlge database command for a transaction or
// snapshot or an iterator. It is defined as a string in the following format:
//
// db:%s new-transaction         => error:nil, tx:%s
// tx:%s get key:%s              => error:ErrNotExist
// tx:%s set key:%s value:%s
// tx:%s get key:%s              => value:%s
// tx:%s delete key:%s
// tx:%s get key:%s              => error:ErrNotExist
// tx:%s commit                  =>
// tx:%s rollback                => error:non-nil
//
// db:%s new-snapshot            => snap:%s
// snap:%s get key:%s            => error:ErrNotExist value:%s
// snap:%s discard
//
// tx:%s ascend begin:%s end:%s    => nil it:%s
//
// it:%s current => key:%s value:%s
// it:%s next
//
// Database commands above are parsed and validated into an object which can
// be used to run the command on a user database.
type TemplateStep struct {
	line string

	// Op holds the test action, which is one of the templateOps defined above.
	Op TemplateOp

	Database    string
	Transaction string
	Snapshot    string
	Iterator    string

	Key   string
	Value string

	Begin string
	End   string

	Error string
}

// TemplateStepResult holds the result of a TemplateStep execution.
type TemplateStepResult struct {
	Step   *TemplateStep
	Status error

	Key   string
	Value io.Reader

	Iterator    kv.Iterator
	Snapshot    kv.Snapshot
	Transaction kv.Transaction
}

// ParseTemplateStep parses input string into a database command.
func ParseTemplateStep(s string) (*TemplateStep, error) {
	step := new(TemplateStep)
	words := strings.Fields(strings.TrimSpace(s))
	for i, word := range words {
		if strings.HasPrefix(word, "#") {
			words = words[:i]
			break
		}

		switch {
		case strings.HasPrefix(word, "db:"):
			step.Database = strings.TrimPrefix(word, "db:")
		case strings.HasPrefix(word, "tx:"):
			step.Transaction = strings.TrimPrefix(word, "tx:")
		case strings.HasPrefix(word, "snap:"):
			step.Snapshot = strings.TrimPrefix(word, "snap:")
		case strings.HasPrefix(word, "it:"):
			step.Iterator = strings.TrimPrefix(word, "it:")
		case strings.HasPrefix(word, "key:"):
			step.Key = strings.TrimPrefix(word, "key:")
		case strings.HasPrefix(word, "value:"):
			step.Value = strings.TrimPrefix(word, "value:")
		case strings.HasPrefix(word, "error:"):
			step.Error = strings.TrimPrefix(word, "error:")
		case strings.HasPrefix(word, "begin:"):
			step.Begin = strings.TrimPrefix(word, "begin:")
		case strings.HasPrefix(word, "end:"):
			step.End = strings.TrimPrefix(word, "end:")
		case word == "=>":
			continue
		default:
			if !slices.Contains(ops, TemplateOp(strings.ToLower(word))) {
				return nil, fmt.Errorf("invalid/unrecognized op %q", word)
			}
			step.Op = TemplateOp(strings.ToLower(word))
		}
	}
	if err := step.check(); err != nil {
		return nil, err
	}
	step.line = strings.Join(words, " ")
	return step, nil
}

func (s *TemplateStep) String() string {
	return fmt.Sprintf("{%s}", s.line)
}

func (s *TemplateStep) check() error {
	if s.Transaction != "" && s.Snapshot != "" {
		return os.ErrInvalid
	}

	// TODO: Add more stricter checks -- presence of unnecessary fields should fail.

	switch s.Op {
	case "get", "set", "delete", "ascend", "descend", "scan":
		if s.Transaction == "" && s.Snapshot == "" {
			return fmt.Errorf("%s needs a snapshot or transaction value", s.Op)
		}
	}

	switch s.Op {
	case "get":
		if s.Key == "" {
			return fmt.Errorf("get needs a non-empty key")
		}
	case "set":
		if s.Key == "" || s.Value == "" || s.Transaction == "" {
			return fmt.Errorf("set needs key, value and tx values")
		}
	case "delete":
		if s.Key == "" || s.Transaction == "" {
			return fmt.Errorf("delete needs key and tx values")
		}
	case "ascend", "descend":
		if s.Iterator == "" && s.Error == "" {
			return fmt.Errorf("%q needs iterator name", s.Op)
		}
	case "scan":
		if s.Iterator == "" {
			return fmt.Errorf("scan needs an iterator name")
		}

	case "current", "next":
		if s.Iterator == "" {
			return fmt.Errorf("%q needs an iterator name", s.Op)
		}

	case "new-transaction":
		if s.Database == "" || s.Transaction == "" {
			return fmt.Errorf("new-transaction needs database and transaction names")
		}
	case "commit", "rollback":
		if s.Transaction == "" {
			return fmt.Errorf("%q need a transaction name", s.Op)
		}
	case "new-snapshot":
		if s.Database == "" || s.Snapshot == "" {
			return fmt.Errorf("new-snapshot needs database and transaction names")
		}
	case "discard":
		if s.Snapshot == "" {
			return fmt.Errorf("discard needs a snapshot name")
		}
	}
	return nil
}

func (s *TemplateStep) checkKey(r *TemplateStepResult) error {
	if s.Op != "current" && s.Op != "next" {
		return nil
	}
	if r.Key != s.Key {
		return fmt.Errorf("step %v: want %q got %q", s, s.Key, r.Key)
	}
	return nil
}

func (s *TemplateStep) checkValue(r *TemplateStepResult) error {
	if s.Op != "get" && s.Op != "current" && s.Op != "next" {
		return nil
	}
	var str string
	if r.Value != nil {
		data, err := io.ReadAll(r.Value)
		if err != nil {
			return err
		}
		str = string(data)
	}

	if s.Value != str {
		return fmt.Errorf("step %v: want %q got %q", s, s.Value, str)
	}
	return nil
}

func (s *TemplateStep) checkStatus(r *TemplateStepResult) error {
	errs := strings.Split(s.Error, "|")

	if r.Status == nil {
		if len(errs) == 0 || s.Error == "" || slices.Contains(errs, "nil") {
			return nil
		}
		return fmt.Errorf("step %v: want %q, got nil", s, s.Error)
	}

	if len(errs) == 0 {
		return fmt.Errorf("step %v: want nil, got %v", s, r.Status)
	}

	if errors.Is(r.Status, os.ErrNotExist) {
		if slices.Contains(errs, "ErrNotExist") {
			return nil
		}
	}
	if errors.Is(r.Status, os.ErrInvalid) {
		if slices.Contains(errs, "ErrInvalid") {
			return nil
		}
	}

	if slices.Contains(errs, "non-nil") {
		return nil
	}
	return fmt.Errorf("step %v: want %q, got %v", s, s.Error, r.Status)
}

// Run executes the database command on a user DB.
func (s *TemplateStep) Run(ctx context.Context, db kv.Database, tx kv.Transaction, snap kv.Snapshot, iter kv.Iterator) (*TemplateStepResult, error) {
	result := &TemplateStepResult{
		Step: s,
	}

	switch s.Op {
	default:
		return nil, fmt.Errorf("invalid/unrecognized op %q", s.Op)

	case "get":
		if len(s.Transaction) != 0 {
			result.Value, result.Status = tx.Get(ctx, s.Key)
		}
		if len(s.Snapshot) != 0 {
			result.Value, result.Status = snap.Get(ctx, s.Key)
		}

	case "set":
		result.Status = tx.Set(ctx, s.Key, strings.NewReader(s.Value))
	case "delete":
		result.Status = tx.Delete(ctx, s.Key)

	case "ascend":
		if len(s.Transaction) != 0 {
			result.Iterator, result.Status = tx.Ascend(ctx, s.Begin, s.End)
			break
		}
		if len(s.Snapshot) != 0 {
			result.Iterator, result.Status = snap.Ascend(ctx, s.Begin, s.End)
			break
		}
		return nil, fmt.Errorf("%q has no tx or snap name", s.Op)

	case "descend":
		if len(s.Transaction) != 0 {
			result.Iterator, result.Status = tx.Descend(ctx, s.Begin, s.End)
			break
		}
		if len(s.Snapshot) != 0 {
			result.Iterator, result.Status = snap.Descend(ctx, s.Begin, s.End)
			break
		}
		return nil, fmt.Errorf("%q has no tx or snap name", s.Op)

	case "scan":
		if len(s.Transaction) != 0 {
			result.Iterator, result.Status = tx.Scan(ctx)
			break
		}
		if len(s.Snapshot) != 0 {
			result.Iterator, result.Status = snap.Scan(ctx)
			break
		}
		return nil, fmt.Errorf("%q has no tx or snap name", s.Op)

	case "current":
		if k, v, ok := iter.Current(ctx); ok {
			result.Key, result.Value, result.Status = k, v, nil
		} else {
			result.Key, result.Value, result.Status = k, v, iter.Err()
		}

	case "next":
		if k, v, ok := iter.Next(ctx); ok {
			result.Key, result.Value, result.Status = k, v, nil
		} else {
			result.Key, result.Value, result.Status = k, v, iter.Err()
		}

	case "new-transaction":
		result.Transaction, result.Status = db.NewTransaction(ctx)

	case "commit":
		result.Status = tx.Commit(ctx)

	case "rollback":
		result.Status = tx.Rollback(ctx)

	case "new-snapshot":
		result.Snapshot, result.Status = db.NewSnapshot(ctx)

	case "discard":
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