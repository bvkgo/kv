// Copyright (c) 2023 BVK Chaitanya

package kvmemdb

import (
	"bufio"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/bvkgo/kv/internal/multival"
)

type gobHeader struct {
	LastTxVersion    int64
	MaxCommitVersion int64
}

type gobValue struct {
	Version int64
	Data    []byte
	Key     string
}

// func (v *gobValue) GobEncode() ([]byte, error) {
// 	return []byte(fmt.Sprintf("%s %08d %s\n", v.Key, v.Version, v.Data)), nil
// }

// func (v *gobValue) GobDecode(bs []byte) error {
// 	_, err := fmt.Sscanf(string(bs), "%s %08d %s\n", &v.Key, &v.Version, &v.Data)
// 	return err
// }

// Backup saves database content to a file.
func Backup(db *DB, file string) error {
	fp, err := os.Create(file)
	if err != nil {
		return err
	}
	defer fp.Close()

	bufp := bufio.NewWriter(fp)
	encoder := gob.NewEncoder(bufp)

	snap, _ := db.NewSnapshot(context.Background())
	s := snap.(*Snapshot)
	defer s.Discard(context.Background())

	db.mu.Lock()
	header := gobHeader{
		LastTxVersion:    db.lastTxVersion,
		MaxCommitVersion: db.maxCommitVersion,
	}
	db.mu.Unlock()

	if err := encoder.Encode(header); err != nil {
		return err
	}

	var status error
	save := func(key string, mval *multival.MultiValue) bool {
		if v, ok := mval.Fetch(s.lastCommitVersion); ok && !v.Deleted {
			gv := &gobValue{
				Key:     key,
				Data:    v.Data,
				Version: v.Version,
			}
			if err := encoder.Encode(gv); err != nil {
				status = err
				return false
			}
		}
		return true
	}
	db.store.Range(save)
	if status != nil {
		return fmt.Errorf("could not complete db scan: %w", status)
	}

	if err := bufp.Flush(); err != nil {
		return err
	}
	if err := fp.Sync(); err != nil {
		return err
	}
	return nil
}

// Restore loads database from a file.
func Restore(file string) (*DB, error) {
	fp, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	db := New()

	dec := gob.NewDecoder(bufio.NewReader(fp))
	header := new(gobHeader)
	if err := dec.Decode(header); err != nil {
		return nil, err
	}

	gv := new(gobValue)
	for err = dec.Decode(gv); err == nil; err = dec.Decode(gv) {
		if db.maxCommitVersion < gv.Version {
			db.maxCommitVersion = gv.Version
		}
		v := &multival.Value{
			Version: gv.Version,
			Data:    gv.Data,
		}
		mv := multival.Append(nil, v)
		db.store.Store(gv.Key, mv)
		*gv = gobValue{}
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	db.lastTxVersion = header.LastTxVersion + 1
	db.maxCommitVersion = header.MaxCommitVersion + 1
	return db, nil
}
