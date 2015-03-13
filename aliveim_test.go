package main

import (
	"io"
	"strings"
	"testing"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func TestParseAlivePost(t *testing.T) {
	var body io.ReadCloser = nopCloser{strings.NewReader(`{"device_id": "abc123", "timeout": 300}`)}
	var ar AliveRequest = parseAlivePost(body)

	if ar.DeviceID != "abc123" || ar.Timeout != 300 {
		t.Fatalf("Expected: DeviceID: %s, Timeout: %d, got DeviceID: %s, Timeout: %d", "abc123", 300, ar.DeviceID, ar.Timeout)
	}
}
