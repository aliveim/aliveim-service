package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

const defaultAddr string = "localhost"
const defaultPort string = "5000"

func notifyDeviceTimerExpired(device_id string) {
	log.Printf("DeviceID: %s expired!\n", device_id)
	delete(timers_map, device_id)
	return
}

func handleAlivePost(rw http.ResponseWriter, request *http.Request) {
	aliverequest, err := parseAlivePost(request.Body)
	if err != nil {
		log.Printf("ERROR: Couldn't parse request -- %v", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
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

func parseAlivePost(body io.ReadCloser) (AliveRequest, error) {
	aliverequest_decoder := json.NewDecoder(body)

	var aliverequest AliveRequest
	err := aliverequest_decoder.Decode(&aliverequest)

	return aliverequest, err
}

func Handlers() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", handleAlivePost).Methods("POST")
	return r
}

func main() {
	port := defaultPort
	addr := defaultAddr

	if nargs := len(os.Args[1:]); nargs > 0 {
		switch nargs {
		case 2:
			port = os.Args[2]
			fallthrough
		case 1:
			addr = os.Args[1]
		default:
			log.Fatal("Too many parameters.")
		}
	}

	log.Printf("Starting AliveIM service on %s and port %s...\n", addr, port)
	if err := http.ListenAndServe(fmt.Sprintf("%s:%s", addr, port), Handlers()); err != nil {
		log.Fatalf("%v", err)
	}
}
