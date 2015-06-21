package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type AliveRequest struct {
	DeviceID string `json:"device_id"`
	Timeout  int32  `json:"timeout"`
}

type DeviceTimer struct {
	DeviceID    string
	DeviceTimer *time.Timer
}

func (timer DeviceTimer) startTimer() {
	<-timer.DeviceTimer.C
	notifyDeviceTimerExpired(timer.DeviceID)
}

var timers_map = make(map[string]DeviceTimer)

func notifyDeviceTimerExpired(device_id string) {
	log.Printf("DeviceID: %s expired!\n", device_id)
	delete(timers_map, device_id)
	return
}

func handleAlivePost(rw http.ResponseWriter, request *http.Request) {
	aliverequest := parseAlivePost(request.Body)
	log.Printf("DeviceID: %s, Timeout: %d\n", aliverequest.DeviceID, aliverequest.Timeout)

	timer, timer_found := timers_map[aliverequest.DeviceID]

	if timer_found {
		timer.DeviceTimer.Reset(time.Millisecond * time.Duration(aliverequest.Timeout))
		rw.WriteHeader(http.StatusOK)
	} else {
		timer := time.NewTimer(time.Millisecond * time.Duration(aliverequest.Timeout))
		device_timer := DeviceTimer{aliverequest.DeviceID, timer}
		timers_map[aliverequest.DeviceID] = device_timer
		go device_timer.startTimer()
		rw.WriteHeader(http.StatusCreated)
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

func Handlers() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", handleAlivePost).Methods("POST")
	return r
}

func main() {
	log.Println("Starting AliveIM service...")
	http.ListenAndServe("localhost:5000", Handlers())
}
