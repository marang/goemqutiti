package proxy

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
)

func handleDBWrites(ctx context.Context) {
	<-ctx.Done()
}

func manageBrokerStatus(ctx context.Context) {
	<-ctx.Done()
}

// Run starts the proxy gRPC server on the provided address. It returns the
// started server and a cleanup function to release resources.
func Run(addr string) (*grpc.Server, func(), error) {
	lf, err := Acquire(LockPath())
	if err != nil {
		return nil, nil, err
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		_ = Release(lf)
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
		_ = Release(lf)
	}

	return srv, cleanup, nil
}
