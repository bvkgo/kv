// Copyright (c) 2023 BVK Chaitanya

package internal

import (
	"bufio"
	"context"
	"fmt"
	"io"
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

// RunTemplateTest runs all database operations in the input template serially
// one-after-another in the order defined in the template.
func RunTemplateTest(ctx context.Context, text string, dbs ...kv.Database) ([]*TemplateStepResult, error) {
	test, err := ParseTemplateTest(strings.NewReader(text))
	if err != nil {
		return nil, err
	}

	rt, err := newTemplateRuntime(test.Steps, dbs...)
	if err != nil {
		return nil, err
	}

	var results []*TemplateStepResult
	for _, s := range test.Steps {
		result, err := rt.RunStep(ctx, s)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}
