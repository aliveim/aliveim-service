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

func (timer DeviceTimer) startTimerAndWait() {
	<-timer.DeviceTimer.C
	notifyDeviceTimerExpired(timer.DeviceID)
}

var timers_map = make(map[string]DeviceTimer)

func notifyDeviceTimerExpired(device_id string) {
	return
}

func handleAlivePost(rw http.ResponseWriter, request *http.Request) {
	aliverequest := parseAlivePost(request.Body)
	log.Printf("DeviceID: %s, Timeout: %d\n", aliverequest.DeviceID, aliverequest.Timeout)

	timer, timer_found := timers_map[aliverequest.DeviceID]

	if timer_found {
		timer.DeviceTimer.Reset(time.Duration(aliverequest.Timeout))
	} else {
		timer := time.NewTimer(time.Duration(aliverequest.Timeout))
		device_timer := DeviceTimer{aliverequest.DeviceID, timer, aliverequest.Timeout}
		timers_map[aliverequest.DeviceID] = device_timer
		device_timer.startTimerAndWait()
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
