package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

var (
	server        *httptest.Server
	reader        io.Reader
	devicePostUrl string
)

func init() {
	server = httptest.NewServer(Handlers())
	devicePostUrl = fmt.Sprintf("%s/", server.URL)
}

func TestParseAlivePost(t *testing.T) {
	var body io.ReadCloser = nopCloser{strings.NewReader(`{"device_id": "abc123", "timeout": 300}`)}
	var ar AliveRequest = parseAlivePost(body)

	if ar.DeviceID != "abc123" || ar.Timeout != 300 {
		t.Fatalf("Expected: DeviceID: %s, Timeout: %d, got DeviceID: %s, Timeout: %d",
			"abc123", 300, ar.DeviceID, ar.Timeout)
	}
}

func TestCreateTimerInsertMapRetrive(t *testing.T) {
	var timers_map = make(map[string]DeviceTimer)
	timer := time.NewTimer(time.Second * 2)
	device_timer := DeviceTimer{"abc123", timer}
	timers_map["abc123"] = device_timer
	my_timer := timers_map["abc123"]

	if my_timer.DeviceID != "abc123" {
		t.Fatalf("Expected: DeviceID: %s, got DeviceID: %s", "abc123", my_timer.DeviceID)
	}
}

func TestDeviceTimerStartTimerTimeout(t *testing.T) {
	timer := time.NewTimer(time.Millisecond * 300)
	device_timer := DeviceTimer{"abc123", timer}
	fmt.Println("Start timer...")
	go device_timer.startTimer()
	fmt.Println("Sleep 100 ms...")
	time.Sleep(time.Millisecond * 100)
	fmt.Println("Sleep 300 ms...")
	time.Sleep(time.Millisecond * 300)
	fmt.Println("Printed after device expiration")
}

func TestPostDevicePayloadEmptyTimersMap(t *testing.T) {
	deviceJson := `{"device_id": "abc123", "timeout": 300}`
	reader = strings.NewReader(deviceJson)                         //Convert string to reader
	request, err := http.NewRequest("POST", devicePostUrl, reader) //Create request with JSON body

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		t.Error(err) //Something is wrong while sending request
	}

	if res.StatusCode != 201 {
		t.Errorf("Success expected: %d", res.StatusCode) //Uh-oh this means our test failed
	}

	timer, timer_found := timers_map["abc123"]
	assert.True(t, timer_found)
	assert.NotNil(t, timer)
}

func TestPostDevicePayloadExistingTimersMap(t *testing.T) {
	timer := time.NewTimer(time.Millisecond * time.Duration(300))
	device_timer := DeviceTimer{"abc123", timer}
	timers_map["abc123"] = device_timer

	deviceJson := `{"device_id": "abc123", "timeout": 300}`
	reader = strings.NewReader(deviceJson)                         //Convert string to reader
	request, err := http.NewRequest("POST", devicePostUrl, reader) //Create request with JSON body

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		t.Error(err) //Something is wrong while sending request
	}

	if res.StatusCode != 200 {
		t.Errorf("Success expected: %d", res.StatusCode) //Uh-oh this means our test failed
	}
}
