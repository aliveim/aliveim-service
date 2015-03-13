package main

import (
	"encoding/json"
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
	log.Printf("DeviceID: %s, Timeout: %d\n", aliverequest.DeviceID, aliverequest.Timeout)

	timer := timers_map[aliverequest.DeviceID]

	if timer == nil {
		timer := time.NewTimer(aliverequest.Timeout)
		device_timer := DeviceTimer{aliverequest.DeviceID, timer, aliverequest.Timeout}
		timers_map[aliverequest.DeviceID] = device_timer
	} else {
		timer.DeviceTimer.Reset(aliverequest.Timeout)
	}
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
	log.Println("Starting AliveIM service...")
	http.HandleFunc("/", handleAlivePost)
	http.ListenAndServe("localhost:5000", nil)
}
