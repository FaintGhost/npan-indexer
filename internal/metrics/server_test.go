package metrics_test

import (
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"npan/internal/metrics"
)

func TestMetricsServer(t *testing.T) {
	reg := metrics.NewRegistry()

	// Use port 0 for ephemeral port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	addr := listener.Addr().String()

	srv := metrics.NewMetricsServer(addr, reg)

	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			t.Errorf("server error: %v", err)
		}
	}()
	defer srv.Close()

	// Wait for server to be ready
	time.Sleep(50 * time.Millisecond)

	client := &http.Client{Timeout: 5 * time.Second}

	// Test /metrics endpoint
	resp, err := client.Get("http://" + addr + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want 200", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Errorf("content-type: got %q, want text/plain", ct)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "go_goroutines") {
		t.Error("body should contain go_goroutines")
	}

	// Test non-metrics path returns 404
	resp2, err := client.Get("http://" + addr + "/foo")
	if err != nil {
		t.Fatalf("GET /foo failed: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusNotFound {
		t.Errorf("/foo status: got %d, want 404", resp2.StatusCode)
	}
}
