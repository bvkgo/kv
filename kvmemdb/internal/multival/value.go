// Copyright (c) 2023 BVK Chaitanya

package multival

import (
	"fmt"
	"slices"
	"strings"
)

type Value struct {
	Version int64
	Data    []byte
	Deleted bool
}

func (v *Value) String() string {
	if v.Deleted {
		return fmt.Sprintf("{version:%d deleted}", v.Version)
	}
	return fmt.Sprintf("{version:%d Data:%s}", v.Version, v.Data)
}

type MultiValue struct {
	values []*Value
}

func (mv *MultiValue) String() string {
	var sb strings.Builder
	sb.WriteRune('[')
	for _, v := range mv.values {
		sb.WriteString(v.String())
	}
	sb.WriteRune(']')
	return sb.String()
}

// Fetch returns the value found at the given version or the closest lower
// version to the given version.
func (mv *MultiValue) Fetch(version int64) (*Value, bool) {
	index, ok := slices.BinarySearchFunc(mv.values, version, findValue)
	if ok {
		return mv.values[index], true
	}

	closest := index - 1
	if closest >= 0 && closest < len(mv.values) {
		return mv.values[closest], true
	}

	return nil, false
}

func findValue(v *Value, version int64) int {
	if v.Version == version {
		return 0
	}
	if v.Version < version {
		return -1
	}
	return 1
}

// Append returns a clone of the input multi-value added with a newer version
// to end. Input value version *must* be larger than all existing versions.
func Append(mv *MultiValue, v *Value) *MultiValue {
	if mv == nil || mv.values == nil {
		return &MultiValue{
			values: []*Value{v},
		}
	}

	if !slices.IsSortedFunc(mv.values, func(a, b *Value) int {
		return int(a.Version - b.Version)
	}) {
		panic("multi-value versions are not in sorted order")
	}

	nvalues := len(mv.values)
	if last := mv.values[nvalues-1]; v.Version <= last.Version {
		panic("append version is not larger than existing multi-value versions")
	}

	newvs := make([]*Value, nvalues+1)
	copy(newvs, mv.values)
	newvs[nvalues] = v

	if !slices.IsSortedFunc(newvs, func(a, b *Value) int {
		return int(a.Version - b.Version)
	}) {
		panic("newer multi-value versions are not in sorted order")
	}

	return &MultiValue{values: newvs}
}

// Compact drops older data before the given version when it is not the only
// version. Returns the same input multi-value if no compaction can be
// performed; otherwise, returns a clone of the input multi-value.
func Compact(mv *MultiValue, minVersion int64) *MultiValue {
	if mv == nil || mv.values == nil {
		return nil
	}
	if len(mv.values) <= 1 {
		return mv
	}

	index, ok := slices.BinarySearchFunc(mv.values, minVersion, findValue)
	if !ok {
		index = index - 1
	}
	if index < 0 {
		return mv
	}

	newvs := slices.DeleteFunc(slices.Clone(mv.values), func(v *Value) bool {
		return v.Version < mv.values[index].Version
	})

	if len(newvs) == len(mv.values) {
		return mv
	}
	return &MultiValue{values: newvs}
}

func (mv *MultiValue) Empty() bool {
	if mv == nil || len(mv.values) == 0 {
		return true
	}
	for _, v := range mv.values {
		if !v.Deleted {
			return false
		}
	}
	return true
}
