package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type AliveRequest struct {
	DeviceID string `json:"device_id"`
	Timeout  int32  `json:"timeout"`
}

type DeviceTimer struct {
	DeviceID      string
	DeviceTimer   *time.Timer
	DeviceTimeout int32
}

var timers_map = make(map[string]DeviceTimer)

func handleAlivePost(rw http.ResponseWriter, request *http.Request) {
	aliverequest := parseAlivePost(request.Body)
	fmt.Printf("DeviceID: %s, Timeout: %d", aliverequest.DeviceID, aliverequest.Timeout)
}

func parseAlivePost(body io.ReadCloser) AliveRequest {
	aliverequest_decoder := json.NewDecoder(body)

	var aliverequest AliveRequest
	err_aliverequest := aliverequest_decoder.Decode(&aliverequest)

	if err_aliverequest != nil {
		log.Fatalf("Error decoding aliverequest: %s", err_aliverequest)
	}

	return aliverequest
}

func main() {
	fmt.Println("Starting AliveIM service...")
	http.HandleFunc("/", handleAlivePost)
	http.ListenAndServe("localhost:5000", nil)
}
