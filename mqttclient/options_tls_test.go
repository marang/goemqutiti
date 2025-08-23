package mqttclient

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
	"sync"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func generateMTLSCerts(t *testing.T) (caPEM, srvPEM, srvKey, cliPEM, cliKey []byte) {
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
	return
}

func startMutualTLSServer(t *testing.T, caPEM, srvPEM, srvKey []byte) (addr string, closeFn func()) {
	t.Helper()
	tlsCert, err := tls.X509KeyPair(srvPEM, srvKey)
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

func TestWithTLSCertificateAuth(t *testing.T) {
	caPEM, srvPEM, srvKey, cliPEM, cliKey := generateMTLSCerts(t)
	addr, closeFn := startMutualTLSServer(t, caPEM, srvPEM, srvKey)
	defer closeFn()

	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.pem")
	cliPath := filepath.Join(dir, "client.pem")
	keyPath := filepath.Join(dir, "client.key")
	os.WriteFile(caPath, caPEM, 0644)
	os.WriteFile(cliPath, cliPEM, 0644)
	os.WriteFile(keyPath, cliKey, 0644)

	opts := mqtt.NewClientOptions()
	opts.AddBroker("ssl://" + addr)
	opts.SetClientID("cid")
	opt, err := WithTLS(true, false, caPath, cliPath, keyPath)
	if err != nil {
		t.Fatalf("WithTLS: %v", err)
	}
	opt(opts)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		t.Fatalf("connect: %v", token.Error())
	}
	client.Disconnect(0)
}
