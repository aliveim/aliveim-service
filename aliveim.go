package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
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

type TimersMap struct {
	Timers map[string]DeviceTimer
	sync.RWMutex
}

func NewTimersMap() *TimersMap {
	return &TimersMap{Timers: make(map[string]DeviceTimer)}
}

func (m *TimersMap) ResetOrSetNew(deviceID string, timeout int32) (DeviceTimer, bool) {
	newTimeout := time.Millisecond * time.Duration(timeout)

	m.Lock()
	deviceTimer, ok := m.Timers[deviceID]
	if ok {
		deviceTimer.DeviceTimer.Reset(newTimeout)
	} else {
		timer := time.NewTimer(newTimeout)
		deviceTimer = DeviceTimer{deviceID, timer}
		m.Timers[deviceID] = deviceTimer
		go deviceTimer.startTimer()
	}
	m.Unlock()

	return deviceTimer, !ok
}

func (m *TimersMap) Delete(deviceID string) {
	m.Lock()
	delete(m.Timers, deviceID)
	m.Unlock()
	return
}

const defaultAddr string = "localhost"
const defaultPort string = "5000"

var timers *TimersMap = NewTimersMap()

func notifyDeviceTimerExpired(deviceID string) {
	log.Printf("DeviceID: %s expired!\n", deviceID)
	timers.Delete(deviceID)
	return
}

func handleAlivePost(rw http.ResponseWriter, request *http.Request) {
	aliverequest, err := parseAlivePost(request.Body)
	if err != nil {
		log.Printf("ERROR: Couldn't parse request -- %v", err)
		return
	}
	log.Printf("DeviceID: %s, Timeout: %d\n", aliverequest.DeviceID, aliverequest.Timeout)

	if _, created := timers.ResetOrSetNew(aliverequest.DeviceID, aliverequest.Timeout); created {
		rw.WriteHeader(http.StatusOK)
	} else {
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
