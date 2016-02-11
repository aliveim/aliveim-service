package main

import (
	"encoding/json"
	"flag"
	"fmt"
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

var (
	servicePort = flag.Int("port", 5000, "Set the service port")
	serviceHost = flag.String("host", "localhost", "Set the service host")
	timersMap   = make(map[string]DeviceTimer)
)

func (timer DeviceTimer) startTimer() {
	<-timer.DeviceTimer.C
	notifyDeviceTimerExpired(timer.DeviceID)
}

func notifyDeviceTimerExpired(device_id string) {
	log.Printf("DeviceID: %s expired!\n", device_id)
	delete(timersMap, device_id)
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

	timer, timerFound := timersMap[aliverequest.DeviceID]

	if timerFound {
		timer.DeviceTimer.Reset(time.Millisecond * time.Duration(aliverequest.Timeout))
		rw.WriteHeader(http.StatusOK)
	} else {
		timer := time.NewTimer(time.Millisecond * time.Duration(aliverequest.Timeout))
		deviceTimer := DeviceTimer{aliverequest.DeviceID, timer}
		timersMap[aliverequest.DeviceID] = deviceTimer
		go deviceTimer.startTimer()
		rw.WriteHeader(http.StatusCreated)
	}
}

func parseAlivePost(body io.ReadCloser) (AliveRequest, error) {
	aliverequestDecoder := json.NewDecoder(body)

	var aliverequest AliveRequest
	err := aliverequestDecoder.Decode(&aliverequest)

	return aliverequest, err
}

func Handlers() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", handleAlivePost).Methods("POST")
	return r
}

func main() {
	flag.Parse()

	log.Printf(
		"Starting AliveIM service on %s and port %d ...\n",
		*serviceHost, *servicePort)

	if err := http.ListenAndServe(
		fmt.Sprintf(
			"%s:%d", *serviceHost, *servicePort),
		Handlers()); err != nil {
		log.Fatalf("%v", err)
	}
}
