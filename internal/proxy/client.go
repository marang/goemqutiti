package proxy

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client wraps a gRPC connection to the proxy server.
type Client struct {
	conn *grpc.ClientConn
}

// Connect dials the proxy server at the provided address and returns a Client.
// The connection uses insecure credentials as it is intended for local use.
func Connect(addr string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("dial proxy: %w", err)
	}
	return &Client{conn: conn}, nil
}

// Close terminates the underlying gRPC connection.
func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// CanWrite reports whether the server grants write access. It currently
// returns true and serves as a placeholder for a future RPC call.
func (c *Client) CanWrite(ctx context.Context) (bool, error) {
	return true, nil
}
