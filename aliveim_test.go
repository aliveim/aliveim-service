package main

import (
	"io"
	"strings"
	"testing"
	"time"
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

func TestCreateTimerInsertMapRetrive(t *testing.T) {
	var timers_map = make(map[string]DeviceTimer)
	timer := time.NewTimer(time.Second * 2)
	device_timer := DeviceTimer{"abc123", timer, 2000}
	timers_map["abc123"] = device_timer
	my_timer := timers_map["abc123"]

	if my_timer.DeviceTimeout != 2000 || my_timer.DeviceID != "abc123" {
		t.Fatalf("Expected: DeviceID: %s, Timeout: %d, got DeviceID: %s, Timeout: %d", "abc123", 2000, my_timer.DeviceID, my_timer.DeviceTimeout)
	}
}
