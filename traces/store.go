package traces

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/marang/emqutiti/proxy"
)

var (
	jsonMarshal = json.Marshal
	proxyAddr   string
)

// SetProxyAddr configures the DB proxy address.
func SetProxyAddr(addr string) { proxyAddr = addr }

func addr() string {
	if proxyAddr != "" {
		return proxyAddr
	}
	return os.Getenv("EMQUTITI_PROXY_ADDR")
}

func tracerAddClient(cl proxy.DBProxyClient, profile, key string, msg TracerMessage) error {
	dbKey := fmt.Sprintf("trace/%s/%s/%020d", key, msg.Topic, msg.Timestamp.UnixNano())
	val, err := jsonMarshal(msg)
	if err != nil {
		return err
	}
	_, err = cl.Write(context.Background(), &proxy.WriteRequest{
		Profile: profile,
		Bucket:  "traces",
		Key:     dbKey,
		Value:   val,
	})
	return err
}

func tracerAdd(profile, key string, msg TracerMessage) error {
	cl, conn, err := proxy.NewClient(addr())
	if err != nil {
		return err
	}
	defer conn.Close()
	return tracerAddClient(cl, profile, key, msg)
}

func tracerMessagesClient(cl proxy.DBProxyClient, profile, key string) ([]TracerMessage, error) {
	prefix := fmt.Sprintf("trace/%s/", key)
	resp, err := cl.Read(context.Background(), &proxy.ReadRequest{
		Profile: profile,
		Bucket:  "traces",
		Key:     prefix,
	})
	if err != nil {
		return nil, err
	}
	msgs := make([]TracerMessage, 0, len(resp.Values))
	for _, v := range resp.Values {
		var m TracerMessage
		if err := json.Unmarshal(v, &m); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func tracerMessages(profile, key string) ([]TracerMessage, error) {
	cl, conn, err := proxy.NewClient(addr())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return tracerMessagesClient(cl, profile, key)
}

func tracerHasData(profile, key string) (bool, error) {
	cl, conn, err := proxy.NewClient(addr())
	if err != nil {
		return false, err
	}
	defer conn.Close()
	prefix := fmt.Sprintf("trace/%s/", key)
	resp, err := cl.Read(context.Background(), &proxy.ReadRequest{
		Profile: profile,
		Bucket:  "traces",
		Key:     prefix,
	})
	if err != nil {
		return false, err
	}
	return len(resp.Values) > 0, nil
}

func tracerClearData(profile, key string) error {
	cl, conn, err := proxy.NewClient(addr())
	if err != nil {
		return err
	}
	defer conn.Close()
	prefix := fmt.Sprintf("trace/%s/", key)
	_, err = cl.Delete(context.Background(), &proxy.DeleteRequest{
		Profile: profile,
		Bucket:  "traces",
		Key:     prefix,
	})
	return err
}
