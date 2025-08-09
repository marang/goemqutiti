package emqutiti

import (
	"io"
	"log"
	"testing"
)

func TestStartProxyStatusLoggerWritesLog(t *testing.T) {
	m, _ := initialModel(nil)
	log.SetOutput(io.MultiWriter(io.Discard, m.logs))
	stop := startProxyStatusLogger("")
	stop()
	if len(m.logs.Lines()) == 0 {
		t.Fatalf("expected log line, got none")
	}
}
