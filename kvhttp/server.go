// Copyright (c) 2023 BVK Chaitanya

package kvhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/bvkgo/kv"
	"github.com/bvkgo/kv/internal/syncmap"
	"github.com/bvkgo/kv/kvhttp/api"
	"github.com/google/uuid"
)

type idLock struct {
	id uuid.UUID
	mu sync.Mutex
}

type server struct {
	db kv.Database

	mux *http.ServeMux

	dbPath, txPath, itPath, snapPath string

	// nameMap holds a mapping from client assigned name to an unique, lockable
	// uuid. Clients refer to iterators, snapshots and txes by their names, which
	// are assigned unique uuids on the server side.
	nameMap syncmap.Map[string, *idLock]

	txMap   syncmap.Map[uuid.UUID, kv.Transaction]
	itMap   syncmap.Map[uuid.UUID, kv.Iterator]
	snapMap syncmap.Map[uuid.UUID, kv.Snapshot]

	// txItersMap and snapItersMap hold slice of iterator names for a tx or
	// snapshot. Multiple iterators can be live simultaneously for a single tx
	// (or snapshot). Iterators are closed automatically when tx or snapshot is
	// done.

	txItersMap   syncmap.Map[uuid.UUID, []string]
	snapItersMap syncmap.Map[uuid.UUID, []string]
}

func Handler(db kv.Database) http.Handler {
	s := &server{
		db:  db,
		mux: http.NewServeMux(),
	}

	s.mux.Handle("/new-tx", httpPostJSONHandler(s.newTransaction))
	s.mux.Handle("/new-transaction", httpPostJSONHandler(s.newTransaction))
	s.mux.Handle("/new-snap", httpPostJSONHandler(s.newSnapshot))
	s.mux.Handle("/new-snapshot", httpPostJSONHandler(s.newSnapshot))

	s.mux.Handle("/tx/get", httpPostJSONHandler(s.get))
	s.mux.Handle("/tx/set", httpPostJSONHandler(s.set))
	s.mux.Handle("/tx/del", httpPostJSONHandler(s.del))
	s.mux.Handle("/tx/delete", httpPostJSONHandler(s.del))
	s.mux.Handle("/tx/ascend", httpPostJSONHandler(s.ascend))
	s.mux.Handle("/tx/descend", httpPostJSONHandler(s.descend))
	s.mux.Handle("/tx/scan", httpPostJSONHandler(s.scan))
	s.mux.Handle("/tx/commit", httpPostJSONHandler(s.commit))
	s.mux.Handle("/tx/rollback", httpPostJSONHandler(s.rollback))

	s.mux.Handle("/snap/get", httpPostJSONHandler(s.get))
	s.mux.Handle("/snap/ascend", httpPostJSONHandler(s.ascend))
	s.mux.Handle("/snap/descend", httpPostJSONHandler(s.descend))
	s.mux.Handle("/snap/scan", httpPostJSONHandler(s.scan))
	s.mux.Handle("/snap/discard", httpPostJSONHandler(s.discard))

	s.mux.Handle("/it/fetch", httpPostJSONHandler(s.fetch))
	return s.mux
}

func (s *server) Close() error {
	s.itMap.Range(func(_ uuid.UUID, it kv.Iterator) bool {
		kv.Close(it)
		return true
	})
	s.snapMap.Range(func(_ uuid.UUID, snap kv.Snapshot) bool {
		snap.Discard(context.Background())
		return true
	})
	s.txMap.Range(func(_ uuid.UUID, tx kv.Transaction) bool {
		tx.Rollback(context.Background())
		return true
	})
	s.itMap = syncmap.Map[uuid.UUID, kv.Iterator]{}
	s.snapMap = syncmap.Map[uuid.UUID, kv.Snapshot]{}
	s.txMap = syncmap.Map[uuid.UUID, kv.Transaction]{}
	return nil
}

func (s *server) LockCreate(name string) (id uuid.UUID, exists bool) {
	n := &idLock{
		id: uuid.New(),
	}
	if v, loaded := s.nameMap.LoadOrStore(name, n); loaded {
		v.mu.Lock()
		return v.id, true
	}
	n.mu.Lock()
	return n.id, false
}

func (s *server) LockExisting(name string) (id uuid.UUID, ok bool) {
	v, ok := s.nameMap.Load(name)
	if !ok {
		return id, false
	}
	v.mu.Lock()
	return v.id, true
}

func (s *server) resolveName(name string) (id uuid.UUID, ok bool) {
	v, ok := s.nameMap.Load(name)
	if !ok {
		return id, false
	}
	return v.id, true
}

func (s *server) deleteName(name string) {
	s.nameMap.Delete(name)
}

func (s *server) Unlock(name string, delete bool) {
	v, ok := s.nameMap.Load(name)
	if !ok {
		return
	}
	if delete {
		s.nameMap.Delete(name)
	}
	v.mu.Unlock()
}

type statusErr struct {
	code int
	err  error
}

func (s *statusErr) Error() string {
	return fmt.Sprintf("status code %d: %s", s.code, s.err.Error())
}

func httpPostJSONHandler[T1 any, T2 any](fun func(context.Context, *url.URL, *T1) (*T2, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Printf("invalid method type")
			http.Error(w, "invalid http method type", http.StatusMethodNotAllowed)
			return
		}
		if v := r.Header.Get("content-type"); !strings.EqualFold(v, "application/json") {
			log.Printf("unsupported content type")
			http.Error(w, "unsupported content type", http.StatusBadRequest)
			return
		}
		data, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("invalid body")
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		var req *T1
		if len(data) > 0 {
			req = new(T1)
			if err := json.Unmarshal(data, req); err != nil {
				log.Printf("bad request payload: %v", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		resp, err := fun(r.Context(), r.URL, req)
		if err != nil {
			if se := new(statusErr); errors.As(err, &se) {
				http.Error(w, se.Error(), se.code)
				return
			}
			if errors.Is(err, os.ErrNotExist) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsbytes, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(jsbytes)
	})
}

func (s *server) newTransaction(ctx context.Context, u *url.URL, req *api.NewTransactionRequest) (*api.NewTransactionResponse, error) {
	id, exists := s.LockCreate(req.Name)
	defer s.Unlock(req.Name, false /* delete */)

	if exists {
		return nil, &statusErr{err: os.ErrExist, code: http.StatusConflict}
	}

	tx, err := s.db.NewTransaction(ctx)
	if err != nil {
		return &api.NewTransactionResponse{Error: error2string(err)}, nil
	}

	s.txMap.Store(id, tx)
	return &api.NewTransactionResponse{}, nil
}

func (s *server) set(ctx context.Context, u *url.URL, req *api.SetRequest) (*api.SetResponse, error) {
	id, ok := s.LockExisting(req.Transaction)
	if !ok {
		return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
	}
	defer s.Unlock(req.Transaction, false /* delete */)

	tx, ok := s.txMap.Load(id)
	if !ok {
		return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
	}

	if err := tx.Set(ctx, req.Key, bytes.NewReader(req.Value)); err != nil {
		return &api.SetResponse{Error: error2string(err)}, nil
	}
	return &api.SetResponse{}, nil
}

func (s *server) del(ctx context.Context, u *url.URL, req *api.DeleteRequest) (*api.DeleteResponse, error) {
	id, ok := s.LockExisting(req.Transaction)
	if !ok {
		return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
	}
	defer s.Unlock(req.Transaction, false /* delete */)

	tx, ok := s.txMap.Load(id)
	if !ok {
		return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
	}

	if err := tx.Delete(ctx, req.Key); err != nil {
		return &api.DeleteResponse{Error: error2string(err)}, nil
	}
	return &api.DeleteResponse{}, nil
}

func (s *server) commit(ctx context.Context, u *url.URL, req *api.CommitRequest) (*api.CommitResponse, error) {
	id, ok := s.LockExisting(req.Transaction)
	if !ok {
		return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
	}
	defer s.Unlock(req.Transaction, true /* delete */)

	tx, ok := s.txMap.Load(id)
	if !ok {
		return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
	}
	s.txMap.Delete(id)

	// Close all iterators and delete the iterator names as well.
	if iters, ok := s.txItersMap.Load(id); ok {
		for _, iter := range iters {
			if id, ok := s.resolveName(iter); ok {
				if it, ok := s.itMap.Load(id); ok {
					kv.Close(it)
				}
				s.itMap.Delete(id)
			}
			s.deleteName(iter)
		}
		s.txItersMap.Delete(id)
	}

	if err := tx.Commit(ctx); err != nil {
		return &api.CommitResponse{Error: error2string(err)}, nil
	}
	return &api.CommitResponse{}, nil
}

func (s *server) rollback(ctx context.Context, u *url.URL, req *api.RollbackRequest) (*api.RollbackResponse, error) {
	id, ok := s.LockExisting(req.Transaction)
	if !ok {
		return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
	}
	defer s.Unlock(req.Transaction, true /* delete */)

	tx, ok := s.txMap.Load(id)
	if !ok {
		return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
	}
	s.txMap.Delete(id)

	if iters, ok := s.txItersMap.Load(id); ok {
		for _, iter := range iters {
			if id, ok := s.resolveName(iter); ok {
				if it, ok := s.itMap.Load(id); ok {
					kv.Close(it)
					s.itMap.Delete(id)
				}
			}
			s.deleteName(iter)
		}
		s.txItersMap.Delete(id)
	}

	if err := tx.Rollback(ctx); err != nil {
		return &api.RollbackResponse{Error: error2string(err)}, nil
	}
	return &api.RollbackResponse{}, nil
}

func (s *server) newSnapshot(ctx context.Context, u *url.URL, req *api.NewSnapshotRequest) (*api.NewSnapshotResponse, error) {
	id, exists := s.LockCreate(req.Name)
	defer s.Unlock(req.Name, false /* delete */)

	if exists {
		return nil, &statusErr{err: os.ErrExist, code: http.StatusConflict}
	}

	snap, err := s.db.NewSnapshot(ctx)
	if err != nil {
		return &api.NewSnapshotResponse{Error: error2string(err)}, nil
	}

	s.snapMap.Store(id, snap)
	return &api.NewSnapshotResponse{}, nil
}

func (s *server) discard(ctx context.Context, u *url.URL, req *api.DiscardRequest) (*api.DiscardResponse, error) {
	id, ok := s.LockExisting(req.Snapshot)
	if !ok {
		return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
	}
	defer s.Unlock(req.Snapshot, true /* delete */)

	snap, ok := s.snapMap.Load(id)
	if !ok {
		return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
	}
	s.snapMap.Delete(id)

	if iters, ok := s.snapItersMap.Load(id); ok {
		for _, iter := range iters {
			if id, ok := s.resolveName(iter); ok {
				if it, ok := s.itMap.Load(id); ok {
					kv.Close(it)
					s.itMap.Delete(id)
				}
			}
			s.deleteName(iter)
		}
		s.snapItersMap.Delete(id)
	}

	if err := snap.Discard(ctx); err != nil {
		return &api.DiscardResponse{Error: error2string(err)}, nil
	}
	return &api.DiscardResponse{}, nil
}

func (s *server) get(ctx context.Context, u *url.URL, req *api.GetRequest) (*api.GetResponse, error) {
	if len(req.Transaction) == 0 && len(req.Snapshot) == 0 {
		return nil, &statusErr{err: os.ErrInvalid, code: http.StatusBadRequest}
	}
	if len(req.Transaction) != 0 && len(req.Snapshot) != 0 {
		return nil, &statusErr{err: os.ErrInvalid, code: http.StatusBadRequest}
	}

	var getter kv.Getter
	if len(req.Transaction) != 0 {
		id, ok := s.LockExisting(req.Transaction)
		if !ok {
			return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
		}
		defer s.Unlock(req.Transaction, false /* delete */)

		tx, ok := s.txMap.Load(id)
		if !ok {
			return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
		}
		getter = tx
	} else {
		id, ok := s.LockExisting(req.Snapshot)
		if !ok {
			return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
		}
		defer s.Unlock(req.Snapshot, false /* delete */)

		snap, ok := s.snapMap.Load(id)
		if !ok {
			return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
		}
		getter = snap
	}

	v, err := getter.Get(ctx, req.Key)
	if err != nil {
		return &api.GetResponse{Error: error2string(err)}, nil
	}
	data, err := io.ReadAll(v)
	if err != nil {
		return nil, err
	}
	return &api.GetResponse{Value: data}, nil
}

func (s *server) ascend(ctx context.Context, u *url.URL, req *api.AscendRequest) (*api.AscendResponse, error) {
	if len(req.Transaction) == 0 && len(req.Snapshot) == 0 {
		return nil, &statusErr{err: os.ErrInvalid, code: http.StatusBadRequest}
	}
	if len(req.Transaction) != 0 && len(req.Snapshot) != 0 {
		return nil, &statusErr{err: os.ErrInvalid, code: http.StatusBadRequest}
	}

	id, exists := s.LockCreate(req.Name)
	defer s.Unlock(req.Name, false /* delete */)
	if exists {
		return nil, &statusErr{err: os.ErrExist, code: http.StatusConflict}
	}

	var ranger kv.Ranger
	var rangerID uuid.UUID
	var rangerItersMap *syncmap.Map[uuid.UUID, []string]
	if len(req.Transaction) != 0 {
		id, ok := s.LockExisting(req.Transaction)
		if !ok {
			return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
		}
		defer s.Unlock(req.Transaction, false /* delete */)

		tx, ok := s.txMap.Load(id)
		if !ok {
			return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
		}
		ranger = tx
		rangerID = id
		rangerItersMap = &s.txItersMap
	} else {
		id, ok := s.LockExisting(req.Snapshot)
		if !ok {
			return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
		}
		defer s.Unlock(req.Snapshot, false /* delete */)

		snap, ok := s.snapMap.Load(id)
		if !ok {
			return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
		}
		ranger = snap
		rangerID = id
		rangerItersMap = &s.snapItersMap
	}

	it, err := ranger.Ascend(ctx, req.Begin, req.End)
	if err != nil {
		return &api.AscendResponse{Error: error2string(err)}, nil
	}

	// save the iterator id in iterators-map and it's name in one of tx's or
	// snapshot's iterators map.
	s.itMap.Store(id, it)
	iters, _ := rangerItersMap.Load(rangerID)
	rangerItersMap.Store(rangerID, append(iters, req.Name))

	return &api.AscendResponse{}, nil
}

func (s *server) descend(ctx context.Context, u *url.URL, req *api.DescendRequest) (*api.DescendResponse, error) {
	if len(req.Transaction) == 0 && len(req.Snapshot) == 0 {
		return nil, &statusErr{err: os.ErrInvalid, code: http.StatusBadRequest}
	}
	if len(req.Transaction) != 0 && len(req.Snapshot) != 0 {
		return nil, &statusErr{err: os.ErrInvalid, code: http.StatusBadRequest}
	}

	id, exists := s.LockCreate(req.Name)
	defer s.Unlock(req.Name, false /* delete */)
	if exists {
		return nil, &statusErr{err: os.ErrExist, code: http.StatusConflict}
	}

	var ranger kv.Ranger
	var rangerID uuid.UUID
	var rangerItersMap *syncmap.Map[uuid.UUID, []string]
	if len(req.Transaction) != 0 {
		id, ok := s.LockExisting(req.Transaction)
		if !ok {
			return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
		}
		defer s.Unlock(req.Transaction, false /* delete */)

		tx, ok := s.txMap.Load(id)
		if !ok {
			return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
		}
		ranger = tx
		rangerID = id
		rangerItersMap = &s.txItersMap
	} else {
		id, ok := s.LockExisting(req.Snapshot)
		if !ok {
			return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
		}
		defer s.Unlock(req.Snapshot, false /* delete */)

		snap, ok := s.snapMap.Load(id)
		if !ok {
			return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
		}
		ranger = snap
		rangerID = id
		rangerItersMap = &s.snapItersMap
	}

	it, err := ranger.Descend(ctx, req.Begin, req.End)
	if err != nil {
		return &api.DescendResponse{Error: error2string(err)}, nil
	}

	// save the iterator id in iterators-map and it's name in one of tx's or
	// snapshot's iterators map.
	s.itMap.Store(id, it)
	iters, _ := rangerItersMap.Load(rangerID)
	rangerItersMap.Store(rangerID, append(iters, req.Name))

	return &api.DescendResponse{}, nil
}

func (s *server) scan(ctx context.Context, u *url.URL, req *api.ScanRequest) (*api.ScanResponse, error) {
	if len(req.Transaction) == 0 && len(req.Snapshot) == 0 {
		return nil, &statusErr{err: os.ErrInvalid, code: http.StatusBadRequest}
	}
	if len(req.Transaction) != 0 && len(req.Snapshot) != 0 {
		return nil, &statusErr{err: os.ErrInvalid, code: http.StatusBadRequest}
	}

	id, exists := s.LockCreate(req.Name)
	defer s.Unlock(req.Name, false /* delete */)
	if exists {
		return nil, &statusErr{err: os.ErrExist, code: http.StatusConflict}
	}

	var scanner kv.Scanner
	var scannerID uuid.UUID
	var scannerItersMap *syncmap.Map[uuid.UUID, []string]
	if len(req.Transaction) != 0 {
		id, ok := s.LockExisting(req.Transaction)
		if !ok {
			return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
		}
		defer s.Unlock(req.Transaction, false /* delete */)

		tx, ok := s.txMap.Load(id)
		if !ok {
			return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
		}
		scanner = tx
		scannerID = id
		scannerItersMap = &s.txItersMap
	} else {
		id, ok := s.LockExisting(req.Snapshot)
		if !ok {
			return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
		}
		defer s.Unlock(req.Snapshot, false /* delete */)

		snap, ok := s.snapMap.Load(id)
		if !ok {
			return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
		}
		scanner = snap
		scannerID = id
		scannerItersMap = &s.snapItersMap
	}

	it, err := scanner.Scan(ctx)
	if err != nil {
		return &api.ScanResponse{Error: error2string(err)}, nil
	}

	// save the iterator id in iterators-map and it's name in one of tx's or
	// snapshot's iterators map.
	s.itMap.Store(id, it)
	iters, _ := scannerItersMap.Load(scannerID)
	scannerItersMap.Store(scannerID, append(iters, req.Name))

	return &api.ScanResponse{}, nil
}

func (s *server) fetch(ctx context.Context, u *url.URL, req *api.FetchRequest) (*api.FetchResponse, error) {
	id, ok := s.LockExisting(req.Iterator)
	if !ok {
		return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
	}
	defer s.Unlock(req.Iterator, false /* delete */)

	it, ok := s.itMap.Load(id)
	if !ok {
		return nil, &statusErr{err: os.ErrNotExist, code: http.StatusNotFound}
	}
	k, v, err := it.Fetch(ctx, req.Next)
	if err == nil {
		data, err := io.ReadAll(v)
		if err != nil {
			return nil, err
		}
		return &api.FetchResponse{Key: k, Value: data}, nil
	}
	return &api.FetchResponse{Error: error2string(err)}, nil
}
