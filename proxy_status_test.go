package emqutiti

import (
	"log"
	"regexp"
	"testing"
)

func TestStartProxyStatusLoggerWritesLog(t *testing.T) {
	m, _ := initialModel(nil)
	log.SetFlags(0)
	log.SetOutput(m.logs)
	stop := startProxyStatusLogger("")
	stop()
	lines := m.logs.Lines()
	if len(lines) == 0 {
		t.Fatalf("expected log line, got none")
	}
	if regexp.MustCompile(`^\d{4}/\d{2}/\d{2}`).MatchString(lines[0]) {
		t.Fatalf("unexpected log prefix: %q", lines[0])
	}
}
