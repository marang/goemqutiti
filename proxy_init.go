package emqutiti

import (
	"log"
	"net"
	"time"

	connections "github.com/marang/emqutiti/connections"
	"github.com/marang/emqutiti/proxy"
)

// realInitProxy ensures a DB proxy is running and returns its address.
// If no proxy is reachable, it starts one on the configured or default address.
func realInitProxy() (string, *proxy.Proxy) {
	addr := connections.LoadProxyAddr()
	if addr == "" {
		addr = proxy.DefaultAddr
	}
	if conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond); err == nil {
		conn.Close()
		return addr, nil
	}
	p, err := proxy.StartProxy(addr)
	if err != nil {
		p, err = proxy.StartProxy("127.0.0.1:0")
		if err != nil {
			log.Printf("proxy start failed: %v", err)
			return "", nil
		}
	}
	addr = p.Addr()
	if err := connections.SaveProxyAddr(addr); err != nil {
		log.Printf("save proxy addr: %v", err)
	}
	return addr, p
}

var initProxy = realInitProxy
