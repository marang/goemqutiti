package proxy

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/marang/emqutiti/internal/files"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Proxy runs a gRPC server mediating database access.
type Proxy struct {
	UnimplementedDBProxyServer
	srv *grpc.Server
	lis net.Listener

	mu  sync.Mutex
	dbs map[string]*badger.DB
}

var (
	proxyMu      sync.Mutex
	proxyRunning bool
)

// StartProxy starts the gRPC proxy on addr. Only one may run at a time.
func StartProxy(addr string) (*Proxy, error) {
	proxyMu.Lock()
	defer proxyMu.Unlock()
	if proxyRunning {
		return nil, fmt.Errorf("proxy already running")
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	p := &Proxy{srv: grpc.NewServer(), lis: lis, dbs: make(map[string]*badger.DB)}
	RegisterDBProxyServer(p.srv, p)
	proxyRunning = true
	go p.srv.Serve(lis)
	return p, nil
}

// Stop stops the proxy and closes all database handles.
func (p *Proxy) Stop() {
	p.srv.GracefulStop()
	p.mu.Lock()
	for _, db := range p.dbs {
		db.Close()
	}
	p.dbs = make(map[string]*badger.DB)
	p.mu.Unlock()
	proxyMu.Lock()
	proxyRunning = false
	proxyMu.Unlock()
}

// Addr returns the listening address.
func (p *Proxy) Addr() string { return p.lis.Addr().String() }

func (p *Proxy) dbKey(profile, bucket string) string {
	return profile + "|" + bucket
}

func (p *Proxy) getDB(profile, bucket string) (*badger.DB, error) {
	key := p.dbKey(profile, bucket)
	p.mu.Lock()
	defer p.mu.Unlock()
	if db, ok := p.dbs[key]; ok {
		return db, nil
	}
	if profile == "" {
		profile = "default"
	}
	path := filepath.Join(files.DataDir(profile), bucket)
	if err := files.EnsureDir(path); err != nil {
		return nil, err
	}
	opts := badger.DefaultOptions(path).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	p.dbs[key] = db
	return db, nil
}

// Write stores a key/value pair.
func (p *Proxy) Write(ctx context.Context, req *WriteRequest) (*WriteResponse, error) {
	db, err := p.getDB(req.GetProfile(), req.GetBucket())
	if err != nil {
		return nil, err
	}
	err = db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(req.GetKey()), req.GetValue())
	})
	if err != nil {
		return nil, err
	}
	return &WriteResponse{}, nil
}

// Read returns all values with the given key prefix.
func (p *Proxy) Read(ctx context.Context, req *ReadRequest) (*ReadResponse, error) {
	db, err := p.getDB(req.GetProfile(), req.GetBucket())
	if err != nil {
		return nil, err
	}
	var vals [][]byte
	err = db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(req.GetKey())
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			v, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			vals = append(vals, v)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ReadResponse{Values: vals}, nil
}

// Delete removes all keys with the given prefix.
func (p *Proxy) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	db, err := p.getDB(req.GetProfile(), req.GetBucket())
	if err != nil {
		return nil, err
	}
	err = db.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(req.GetKey())
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if err := txn.Delete(it.Item().KeyCopy(nil)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &DeleteResponse{}, nil
}

// NewClient returns a client connected to the proxy at addr. Used in tests.
func NewClient(addr string) (DBProxyClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return NewDBProxyClient(conn), conn, nil
}
