package proxy

import (
	"context"
	"testing"
)

func TestWriteRead(t *testing.T) {
	p, err := StartProxy("127.0.0.1:0")
	if err != nil {
		t.Fatalf("start proxy: %v", err)
	}
	defer p.Stop()
	client, conn, err := NewClient(p.Addr())
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer conn.Close()
	val := []byte("hello")
	if _, err := client.Write(context.Background(), &WriteRequest{Profile: "p1", Bucket: "b1", Key: "k", Value: val}); err != nil {
		t.Fatalf("write: %v", err)
	}
	resp, err := client.Read(context.Background(), &ReadRequest{Profile: "p1", Bucket: "b1", Key: "k"})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(resp.GetValues()) != 1 || string(resp.GetValues()[0]) != "hello" {
		t.Fatalf("unexpected resp: %v", resp.GetValues())
	}
	if _, err := client.Delete(context.Background(), &DeleteRequest{Profile: "p1", Bucket: "b1", Key: "k"}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	resp, err = client.Read(context.Background(), &ReadRequest{Profile: "p1", Bucket: "b1", Key: "k"})
	if err != nil {
		t.Fatalf("read after delete: %v", err)
	}
	if len(resp.GetValues()) != 0 {
		t.Fatalf("expected empty resp, got %v", resp.GetValues())
	}

	st, err := client.Status(context.Background(), &StatusRequest{})
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if st.GetWrites() == 0 || st.GetReads() == 0 || st.GetDeletes() == 0 {
		t.Fatalf("unexpected counters: %+v", st)
	}
	if st.GetClients() < 1 {
		t.Fatalf("expected at least one client, got %d", st.GetClients())
	}
	if len(st.GetDbs()) == 0 {
		t.Fatalf("expected db info, got %+v", st.GetDbs())
	}
}

func TestProxySingleInstance(t *testing.T) {
	p, err := StartProxy("127.0.0.1:0")
	if err != nil {
		t.Fatalf("start proxy: %v", err)
	}
	defer p.Stop()
	if _, err := StartProxy("127.0.0.1:0"); err == nil {
		t.Fatalf("expected error starting second proxy")
	}
}
