package emqutiti

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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
	var wg sync.WaitGroup
	done := make(chan struct{})
	handshake := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := ln.Accept()
		if err != nil {
			close(handshake)
			return
		}
		defer conn.Close()
		// Read some of the CONNECT packet then respond with CONNACK success.
		buf := make([]byte, 1024)
		conn.Read(buf)
		conn.Write([]byte{0x20, 0x02, 0x00, 0x00})
		close(handshake)
		<-done
	}()
	return ln.Addr().String(), func() {
		ln.Close()
		<-handshake
		close(done)
		wg.Wait()
	}
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

func generateMTLSCerts(t *testing.T) (caPEM, srvPEM, srvKey []byte, cliPEM, cliKey []byte, badCliPEM, badCliKey []byte) {
	t.Helper()
	caPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate ca key: %v", err)
	}
	caTmpl := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, &caTmpl, &caTmpl, &caPriv.PublicKey, caPriv)
	if err != nil {
		t.Fatalf("create ca cert: %v", err)
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatalf("parse ca cert: %v", err)
	}
	caPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})

	srvPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate server key: %v", err)
	}
	srvTmpl := x509.Certificate{
		SerialNumber: big.NewInt(2),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"127.0.0.1"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	srvDER, err := x509.CreateCertificate(rand.Reader, &srvTmpl, caCert, &srvPriv.PublicKey, caPriv)
	if err != nil {
		t.Fatalf("create server cert: %v", err)
	}
	srvPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srvDER})
	srvKey = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(srvPriv)})

	cliPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate client key: %v", err)
	}
	cliTmpl := x509.Certificate{
		SerialNumber: big.NewInt(3),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	cliDER, err := x509.CreateCertificate(rand.Reader, &cliTmpl, caCert, &cliPriv.PublicKey, caPriv)
	if err != nil {
		t.Fatalf("create client cert: %v", err)
	}
	cliPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cliDER})
	cliKey = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(cliPriv)})

	badPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate bad key: %v", err)
	}
	badTmpl := x509.Certificate{
		SerialNumber: big.NewInt(4),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	badDER, err := x509.CreateCertificate(rand.Reader, &badTmpl, &badTmpl, &badPriv.PublicKey, badPriv)
	if err != nil {
		t.Fatalf("create bad client cert: %v", err)
	}
	badCliPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: badDER})
	badCliKey = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(badPriv)})
	return
}

func startMutualTLSServer(t *testing.T, caPEM, certPEM, keyPEM []byte) (addr string, closeFn func()) {
	t.Helper()
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("x509 key pair: %v", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		t.Fatalf("append ca cert")
	}
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
	})
	if err != nil {
		t.Fatalf("tls listen: %v", err)
	}
	var wg sync.WaitGroup
	done := make(chan struct{})
	handshake := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := ln.Accept()
		if err != nil {
			close(handshake)
			return
		}
		defer conn.Close()
		buf := make([]byte, 1024)
		conn.Read(buf)
		conn.Write([]byte{0x20, 0x02, 0x00, 0x00})
		close(handshake)
		<-done
	}()
	return ln.Addr().String(), func() {
		ln.Close()
		<-handshake
		close(done)
		wg.Wait()
	}
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

func TestNewMQTTClientTLSWithCerts(t *testing.T) {
	caPEM, srvPEM, srvKey, cliPEM, cliKey, badCliPEM, badCliKey := generateMTLSCerts(t)
	addr, closeFn := startMutualTLSServer(t, caPEM, srvPEM, srvKey)
	defer closeFn()
	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)
	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.pem")
	cliPath := filepath.Join(dir, "client.pem")
	keyPath := filepath.Join(dir, "client.key")
	os.WriteFile(caPath, caPEM, 0644)
	os.WriteFile(cliPath, cliPEM, 0644)
	os.WriteFile(keyPath, cliKey, 0644)
	p := connections.Profile{Schema: "ssl", Host: host, Port: port, SSL: true, ClientID: "cid", CACertPath: caPath, ClientCertPath: cliPath, ClientKeyPath: keyPath}
	c, err := NewMQTTClient(p, nil)
	if err != nil {
		t.Fatalf("NewMQTTClient: %v", err)
	}
	c.Disconnect()

	badCertPath := filepath.Join(dir, "bad.pem")
	badKeyPath := filepath.Join(dir, "bad.key")
	os.WriteFile(badCertPath, badCliPEM, 0644)
	os.WriteFile(badKeyPath, badCliKey, 0644)
	p.ClientCertPath = badCertPath
	p.ClientKeyPath = badKeyPath
	if _, err := NewMQTTClient(p, nil); err == nil {
		t.Fatalf("expected error for bad client cert")
	}
}
