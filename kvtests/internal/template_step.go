// Copyright (c) 2023 BVK Chaitanya

package internal

import (
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
	"ok",

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

	OK    string
	Key   string
	Value string

	Begin string
	End   string

	Error string

	prefixMap map[string]struct{}
}

// TemplateStepResult holds the result of a TemplateStep execution.
type TemplateStepResult struct {
	Step   *TemplateStep
	Status error

	OK    bool
	Key   string
	Value io.Reader

	Iterator    kv.Iterator
	Snapshot    kv.Snapshot
	Transaction kv.Transaction
}

// ParseTemplateStep parses input string into a database command.
func ParseTemplateStep(s string) (*TemplateStep, error) {
	step := &TemplateStep{
		prefixMap: make(map[string]struct{}),
	}
	words := strings.Fields(strings.TrimSpace(s))
	for i, word := range words {
		if strings.HasPrefix(word, "#") {
			words = words[:i]
			break
		}

		switch {
		case strings.HasPrefix(word, "db:"):
			step.Database = strings.TrimPrefix(word, "db:")
			step.prefixMap["db"] = struct{}{}

		case strings.HasPrefix(word, "tx:"):
			step.Transaction = strings.TrimPrefix(word, "tx:")
			step.prefixMap["tx"] = struct{}{}

		case strings.HasPrefix(word, "snap:"):
			step.Snapshot = strings.TrimPrefix(word, "snap:")
			step.prefixMap["snap"] = struct{}{}

		case strings.HasPrefix(word, "it:"):
			step.Iterator = strings.TrimPrefix(word, "it:")
			step.prefixMap["it"] = struct{}{}

		case strings.HasPrefix(word, "key:"):
			step.Key = strings.TrimPrefix(word, "key:")
			step.prefixMap["key"] = struct{}{}

		case strings.HasPrefix(word, "value:"):
			step.Value = strings.TrimPrefix(word, "value:")
			step.prefixMap["value"] = struct{}{}

		case strings.HasPrefix(word, "ok:"):
			step.OK = strings.TrimPrefix(word, "ok:")
			step.prefixMap["ok"] = struct{}{}

		case strings.HasPrefix(word, "error:"):
			step.Error = strings.TrimPrefix(word, "error:")
			step.prefixMap["error"] = struct{}{}

		case strings.HasPrefix(word, "begin:"):
			step.Begin = strings.TrimPrefix(word, "begin:")
			step.prefixMap["begin"] = struct{}{}

		case strings.HasPrefix(word, "end:"):
			step.End = strings.TrimPrefix(word, "end:")
			step.prefixMap["end"] = struct{}{}

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

func (s *TemplateStep) checkOK(r *TemplateStepResult) error {
	if s.Op != "current" && s.Op != "next" {
		return nil
	}
	if _, ok := s.prefixMap["ok"]; !ok {
		return nil
	}
	if s.OK == "true" && !r.OK {
		return fmt.Errorf("step %v: want %q got %b", s, s.OK, r.OK)
	} else if s.OK == "false" && r.OK {
		return fmt.Errorf("step %v: want %q got %b", s, s.OK, r.OK)
	}
	return nil
}

func (s *TemplateStep) checkKey(r *TemplateStepResult) error {
	if s.Op != "current" && s.Op != "next" {
		return nil
	}
	if _, ok := s.prefixMap["key"]; !ok {
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
	if _, ok := s.prefixMap["value"]; !ok {
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
