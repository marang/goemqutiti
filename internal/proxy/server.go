package proxy

import (
	"context"
	"log"
	"net"
	"os"
	"path/filepath"
	"syscall"

	"google.golang.org/grpc"
)

// acquireLock ensures that only one proxy instance is running. It returns a
// function that releases the lock when called.
func acquireLock() (func(), error) {
	lockPath := filepath.Join(os.TempDir(), "emqutiti-proxy.lock")
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, err
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
		_ = os.Remove(lockPath)
	}, nil
}

func handleDBWrites(ctx context.Context) {
	<-ctx.Done()
}

func manageBrokerStatus(ctx context.Context) {
	<-ctx.Done()
}

// Run starts the proxy gRPC server on the provided address. It returns the
// started server and a cleanup function to release resources.
func Run(addr string) (*grpc.Server, func(), error) {
	unlock, err := acquireLock()
	if err != nil {
		return nil, nil, err
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		unlock()
		return nil, nil, err
	}

	srv := grpc.NewServer()
	// Register generated protobuf services with srv here when available.

	ctx, cancel := context.WithCancel(context.Background())
	go handleDBWrites(ctx)
	go manageBrokerStatus(ctx)

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Printf("proxy gRPC server stopped: %v", err)
		}
	}()

	cleanup := func() {
		cancel()
		unlock()
	}

	return srv, cleanup, nil
}
