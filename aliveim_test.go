package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	devicePostURL string
)

func testTools(code int, body string) (*httptest.Server, *Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, body)
	}))

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	httpClient := &http.Client{Transport: transport}
	mockedClient := &Client{httpClient}

	return server, mockedClient
}

func init() {
	server = httptest.NewServer(Handlers())
	devicePostURL = fmt.Sprintf("%s/", server.URL)
	_, client = testTools(200, `{"status": "ok"}`)
}

func TestParseAlivePost(t *testing.T) {
	var body io.ReadCloser = nopCloser{strings.NewReader(`{"device_id": "abc123", "timeout": 300}`)}
	ar, _ := parseAlivePost(body)

	if ar.DeviceID != "abc123" || ar.Timeout != 300 {
		t.Fatalf("Expected: DeviceID: %s, Timeout: %d, got DeviceID: %s, Timeout: %d",
			"abc123", 300, ar.DeviceID, ar.Timeout)
	}
}

func TestCreateTimerInsertMapRetrive(t *testing.T) {
	var timersMap = make(map[string]DeviceTimer)
	timer := time.NewTimer(time.Second * 2)
	deviceTimer := DeviceTimer{"abc123", timer}
	timersMap["abc123"] = deviceTimer
	myTimer := timersMap["abc123"]

	if myTimer.DeviceID != "abc123" {
		t.Fatalf("Expected: DeviceID: %s, got DeviceID: %s", "abc123", myTimer.DeviceID)
	}
}

func TestDeviceTimerStartTimerTimeout(t *testing.T) {
	timer := time.NewTimer(time.Millisecond * 300)
	deviceTimer := DeviceTimer{"abc123", timer}
	fmt.Println("Start timer...")
	go deviceTimer.startTimer()
	fmt.Println("Sleep 100 ms...")
	time.Sleep(time.Millisecond * 100)
	fmt.Println("Sleep 300 ms...")
	time.Sleep(time.Millisecond * 300)
	fmt.Println("Printed after device expiration")
}

func TestPostDevicePayloadEmptyTimersMap(t *testing.T) {
	deviceJSON := `{"device_id": "abc123", "timeout": 300}`
	reader = strings.NewReader(deviceJSON)                         //Convert string to reader
	request, err := http.NewRequest("POST", devicePostURL, reader) //Create request with JSON body

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		t.Error(err) //Something is wrong while sending request
	}

	if res.StatusCode != 201 {
		t.Errorf("Success expected: %d", res.StatusCode) //Uh-oh this means our test failed
	}

	timer, timerFound := timersMap["abc123"]
	assert.True(t, timerFound)
	assert.NotNil(t, timer)
}

func TestPostDevicePayloadExistingTimersMap(t *testing.T) {
	timer := time.NewTimer(time.Millisecond * time.Duration(300))
	deviceTimer := DeviceTimer{"abc123", timer}
	timersMap["abc123"] = deviceTimer

	deviceJSON := `{"device_id": "abc123", "timeout": 300}`
	reader = strings.NewReader(deviceJSON)                         //Convert string to reader
	request, err := http.NewRequest("POST", devicePostURL, reader) //Create request with JSON body

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		t.Error(err) //Something is wrong while sending request
	}

	if res.StatusCode != 200 {
		t.Errorf("Success expected: %d", res.StatusCode) //Uh-oh this means our test failed
	}
}

func TestMalformedJSONPayLoad(t *testing.T) {
	reader := strings.NewReader("") // empty request
	request, err := http.NewRequest("POST", devicePostURL, reader)
	res, err := http.DefaultClient.Do(request)

	if err != nil {
		t.Error(err) //Something is wrong while sending request
	}
	if res.StatusCode != 400 {
		t.Errorf("Failure expected: %d", res.StatusCode) //Uh-oh this means our test failed
	}
}

func TestNotifyAPIDeviceTimerExpiredSuccess(t *testing.T) {
	server, client = testTools(200, `{"status": "ok"}`)
	defer server.Close()

	err := client.notifyAPIDeviceTimerExpired("abcd1234")
	assert.Nil(t, err)
}
