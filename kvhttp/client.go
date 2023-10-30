// Copyright (c) 2023 BVK Chaitanya

package kvhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/bvkgo/kv"
	"github.com/bvkgo/kv/kvhttp/api"
	"github.com/google/uuid"
)

type DB struct {
	dbURL url.URL

	httpClient *http.Client

	closecalls []func()
}

type Tx struct {
	db *DB
	id string
}

type Snap struct {
	db *DB
	id string
}

type Iter struct {
	db *DB
	id string

	cache struct {
		err   error
		ok    bool
		key   string
		value io.Reader
	}
}

func New(baseURL *url.URL, client *http.Client) *DB {
	if client == nil {
		client = http.DefaultClient
	}

	db := &DB{
		httpClient: client,
		dbURL: url.URL{
			Host:   baseURL.Host,
			Scheme: baseURL.Scheme,
			Path:   baseURL.Path,
		},
	}
	return db
}

func (db *DB) Close() error {
	return nil
}

func (db *DB) ServerURL() url.URL {
	return db.dbURL
}

func (db *DB) NewTransaction(ctx context.Context) (kv.Transaction, error) {
	id := uuid.New().String()
	resp, err := doPost[api.NewTransactionResponse](ctx, db, "/new-transaction", &api.NewTransactionRequest{Name: id})
	if err != nil {
		return nil, err
	}
	if len(resp.Error) != 0 {
		return nil, string2error(resp.Error)
	}
	return &Tx{db: db, id: id}, nil
}

func (db *DB) NewSnapshot(ctx context.Context) (kv.Snapshot, error) {
	id := uuid.New().String()
	resp, err := doPost[api.NewSnapshotResponse](ctx, db, "/new-snapshot", &api.NewSnapshotRequest{Name: id})
	if err != nil {
		return nil, err
	}
	if len(resp.Error) != 0 {
		return nil, string2error(resp.Error)
	}
	return &Snap{db: db, id: id}, nil
}

func (tx *Tx) Get(ctx context.Context, key string) (io.Reader, error) {
	req := &api.GetRequest{Transaction: tx.id, Key: key}
	resp, err := doPost[api.GetResponse](ctx, tx.db, "/tx/get", req)
	if err != nil {
		return nil, err
	}
	if len(resp.Error) != 0 {
		return nil, string2error(resp.Error)
	}
	return bytes.NewReader(resp.Value), nil
}

func (tx *Tx) Set(ctx context.Context, key string, value io.Reader) error {
	data, err := io.ReadAll(value)
	if err != nil {
		return err
	}
	req := &api.SetRequest{
		Transaction: tx.id,
		Key:         key,
		Value:       data,
	}
	resp, err := doPost[api.SetResponse](ctx, tx.db, "/tx/set", req)
	if err != nil {
		return err
	}
	if len(resp.Error) != 0 {
		return string2error(resp.Error)
	}
	return nil
}

func (tx *Tx) Delete(ctx context.Context, key string) error {
	req := &api.DeleteRequest{Transaction: tx.id, Key: key}
	resp, err := doPost[api.DeleteResponse](ctx, tx.db, "/tx/delete", req)
	if err != nil {
		return err
	}
	if len(resp.Error) != 0 {
		return string2error(resp.Error)
	}
	return nil
}

func (tx *Tx) Ascend(ctx context.Context, begin, end string) (kv.Iterator, error) {
	req := &api.AscendRequest{
		Transaction: tx.id,
		Name:        uuid.New().String(),
		Begin:       begin,
		End:         end,
	}
	resp, err := doPost[api.AscendResponse](ctx, tx.db, "/tx/ascend", req)
	if err != nil {
		return nil, err
	}
	if len(resp.Error) != 0 {
		return nil, string2error(resp.Error)
	}
	it := &Iter{db: tx.db, id: req.Name}
	if err := it.init(ctx); err != nil {
		return nil, err
	}
	return it, nil
}

func (tx *Tx) Descend(ctx context.Context, begin, end string) (kv.Iterator, error) {
	req := &api.DescendRequest{
		Transaction: tx.id,
		Name:        uuid.New().String(),
		Begin:       begin,
		End:         end,
	}
	resp, err := doPost[api.DescendResponse](ctx, tx.db, "/tx/descend", req)
	if err != nil {
		return nil, err
	}
	if len(resp.Error) != 0 {
		return nil, string2error(resp.Error)
	}
	it := &Iter{db: tx.db, id: req.Name}
	if err := it.init(ctx); err != nil {
		return nil, err
	}
	return it, nil
}

func (tx *Tx) Scan(ctx context.Context) (kv.Iterator, error) {
	req := &api.ScanRequest{Transaction: tx.id, Name: uuid.New().String()}
	resp, err := doPost[api.ScanResponse](ctx, tx.db, "/tx/scan", req)
	if err != nil {
		return nil, err
	}
	if len(resp.Error) != 0 {
		return nil, string2error(resp.Error)
	}
	it := &Iter{db: tx.db, id: req.Name}
	if err := it.init(ctx); err != nil {
		return nil, err
	}
	return it, nil
}

func (tx *Tx) Commit(ctx context.Context) error {
	req := &api.CommitRequest{Transaction: tx.id}
	resp, err := doPost[api.CommitResponse](ctx, tx.db, "/tx/commit", req)
	if err != nil {
		return err
	}
	if len(resp.Error) != 0 {
		return string2error(resp.Error)
	}
	return nil
}

func (tx *Tx) Rollback(ctx context.Context) error {
	req := &api.RollbackRequest{Transaction: tx.id}
	resp, err := doPost[api.RollbackResponse](ctx, tx.db, "/tx/rollback", req)
	if err != nil {
		return err
	}
	if len(resp.Error) != 0 {
		return string2error(resp.Error)
	}
	return nil
}

func (snap *Snap) Get(ctx context.Context, key string) (io.Reader, error) {
	req := &api.GetRequest{Snapshot: snap.id, Key: key}
	resp, err := doPost[api.GetResponse](ctx, snap.db, "/snap/get", req)
	if err != nil {
		return nil, err
	}
	if len(resp.Error) != 0 {
		return nil, string2error(resp.Error)
	}
	return bytes.NewReader(resp.Value), nil
}

func (snap *Snap) Ascend(ctx context.Context, begin, end string) (kv.Iterator, error) {
	req := &api.AscendRequest{
		Snapshot: snap.id,
		Name:     uuid.New().String(),
		Begin:    begin,
		End:      end,
	}
	resp, err := doPost[api.AscendResponse](ctx, snap.db, "/snap/ascend", req)
	if err != nil {
		return nil, err
	}
	if len(resp.Error) != 0 {
		return nil, string2error(resp.Error)
	}
	it := &Iter{db: snap.db, id: req.Name}
	if err := it.init(ctx); err != nil {
		return nil, err
	}
	return it, nil
}

func (snap *Snap) Descend(ctx context.Context, begin, end string) (kv.Iterator, error) {
	req := &api.DescendRequest{
		Snapshot: snap.id,
		Name:     uuid.New().String(),
		Begin:    begin,
		End:      end,
	}
	resp, err := doPost[api.DescendResponse](ctx, snap.db, "/snap/descend", req)
	if err != nil {
		return nil, err
	}
	if len(resp.Error) != 0 {
		return nil, string2error(resp.Error)
	}
	it := &Iter{db: snap.db, id: req.Name}
	if err := it.init(ctx); err != nil {
		return nil, err
	}
	return it, nil
}

func (snap *Snap) Scan(ctx context.Context) (kv.Iterator, error) {
	req := &api.ScanRequest{Snapshot: snap.id, Name: uuid.New().String()}
	resp, err := doPost[api.ScanResponse](ctx, snap.db, "/snap/scan", req)
	if err != nil {
		return nil, err
	}
	if len(resp.Error) != 0 {
		return nil, string2error(resp.Error)
	}
	it := &Iter{db: snap.db, id: req.Name}
	if err := it.init(ctx); err != nil {
		return nil, err
	}
	return it, nil
}

func (snap *Snap) Discard(ctx context.Context) error {
	req := &api.DiscardRequest{Snapshot: snap.id}
	resp, err := doPost[api.DiscardResponse](ctx, snap.db, "/snap/discard", req)
	if err != nil {
		return err
	}
	if len(resp.Error) != 0 {
		return string2error(resp.Error)
	}
	return nil
}

func (it *Iter) init(ctx context.Context) error {
	req := &api.CurrentRequest{Iterator: it.id}
	resp, err := doPost[api.NextResponse](ctx, it.db, "/it/current", req)
	if err != nil {
		return err
	}
	if len(resp.Error) != 0 {
		return string2error(resp.Error)
	}
	if !resp.OK {
		it.cache.err = io.EOF
		return nil
	}
	it.cache.ok = true
	it.cache.key = resp.Key
	it.cache.value = bytes.NewReader(resp.Value)
	return nil
}

func (it *Iter) Err() error {
	if !errors.Is(it.cache.err, io.EOF) {
		return it.cache.err
	}
	return nil
}

func (it *Iter) Current(ctx context.Context) (string, io.Reader, bool) {
	if it.cache.err != nil {
		return "", nil, false
	}
	return it.cache.key, it.cache.value, it.cache.ok
}

func (it *Iter) Next(ctx context.Context) (string, io.Reader, bool) {
	if it.cache.err != nil {
		return "", nil, false
	}
	req := &api.NextRequest{Iterator: it.id}
	resp, err := doPost[api.NextResponse](ctx, it.db, "/it/next", req)
	if err != nil {
		it.cache.err = err
		return "", nil, false
	}
	if len(resp.Error) != 0 {
		it.cache.err = string2error(resp.Error)
		return "", nil, false
	}
	if !resp.OK {
		it.cache.err = io.EOF
		return "", nil, false
	}
	it.cache.ok = true
	it.cache.key = resp.Key
	it.cache.value = bytes.NewReader(resp.Value)
	return it.cache.key, it.cache.value, it.cache.ok
}

func doPost[RESP, REQ any](ctx context.Context, db *DB, subpath string, req *REQ) (*RESP, error) {
	u := url.URL{
		Host:   db.dbURL.Host,
		Scheme: db.dbURL.Scheme,
		Path:   path.Join(db.dbURL.Path, subpath),
	}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	r.Header.Set("content-type", "application/json")
	resp, err := db.httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-ok http status %d", resp.StatusCode)
	}
	response := new(RESP)
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return nil, err
	}
	return response, nil
}
