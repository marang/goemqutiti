package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/marang/emqutiti/internal/proxy"
)

func main() {
	srv, cleanup, err := proxy.Run(":50051")
	if err != nil {
		log.Fatalf("start proxy: %v", err)
	}
	defer cleanup()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	srv.GracefulStop()
}
