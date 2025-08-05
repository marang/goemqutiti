package emqutiti

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	connections "github.com/marang/emqutiti/connections"
)

// startTLSServer starts a minimal TLS server that speaks enough MQTT to accept a connection.
func startTLSServer(t *testing.T) (addr string, closeFn func()) {
	t.Helper()
	cert, key := generateCert(t)
	tlsCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		t.Fatalf("x509 key pair: %v", err)
	}
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{tlsCert}})
	if err != nil {
		t.Fatalf("tls listen: %v", err)
	}
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		// Read some of the CONNECT packet then respond with CONNACK success.
		buf := make([]byte, 1024)
		conn.Read(buf)
		conn.Write([]byte{0x20, 0x02, 0x00, 0x00})
		time.Sleep(100 * time.Millisecond)
	}()
	return ln.Addr().String(), func() { ln.Close() }
}

func generateCert(t *testing.T) (certPEM, keyPEM []byte) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"127.0.0.1"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	return
}

func TestNewMQTTClientTLSSkipVerify(t *testing.T) {
	addr, closeFn := startTLSServer(t)
	defer closeFn()
	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)
	p := connections.Profile{Schema: "ssl", Host: host, Port: port, SSL: true, SkipTLSVerify: true, ClientID: "cid"}
	c, err := NewMQTTClient(p, nil)
	if err != nil {
		t.Fatalf("NewMQTTClient: %v", err)
	}
	defer c.Disconnect()
	or := c.Client.OptionsReader()
	if !or.TLSConfig().InsecureSkipVerify {
		t.Fatalf("expected InsecureSkipVerify true")
	}
}

func TestNewMQTTClientTLSSecure(t *testing.T) {
	addr, closeFn := startTLSServer(t)
	defer closeFn()
	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)
	p := connections.Profile{Schema: "ssl", Host: host, Port: port, SSL: true, ClientID: "cid"}
	if _, err := NewMQTTClient(p, nil); err == nil || !strings.Contains(err.Error(), "certificate") {
		t.Fatalf("expected cert error, got %v", err)
	}
}
