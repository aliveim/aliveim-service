package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type Client struct {
	HTTPClient *http.Client
}

type AliveRequest struct {
	DeviceID string `json:"device_id"`
	Timeout  int32  `json:"timeout"`
}

type DeviceTimer struct {
	DeviceID    string
	DeviceTimer *time.Timer
}

var (
	servicePort = flag.Int("service-port", 5000, "Set the service port")
	serviceHost = flag.String("service-host", "localhost", "Set the service host")
	apiURL      = flag.String("api-url", "http://localhost", "Set the API url")
	apiTOKEN    = flag.String("api-token", "aabbccdd", "Set the API token")
	timersMap   = make(map[string]DeviceTimer)
	mutex       = &sync.Mutex{}
	client      = &Client{}
)

func (timer DeviceTimer) startTimer() {
	<-timer.DeviceTimer.C
	notifyDeviceTimerExpired(timer.DeviceID)
}

func notifyDeviceTimerExpired(deviceID string) {
	log.Printf("DeviceID: %s expired!\n", deviceID)
	mutex.Lock()
	delete(timersMap, deviceID)
	mutex.Unlock()

	err := client.notifyAPIDeviceTimerExpired(deviceID)

	if err != nil {
		log.Printf("Error while posting notification to the API server: %s\n", err)
	}
}

func (m *Client) notifyAPIDeviceTimerExpired(deviceID string) (err error) {
	var jsonStr = []byte(fmt.Sprintf(`{"device_id": %s}`, deviceID))
	req, err := http.NewRequest("POST", *apiURL, bytes.NewBuffer(jsonStr))

	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Token %s", *apiTOKEN))
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.HTTPClient.Do(req)
	if err != nil {
		fmt.Println("Error POSTing on API server")
		resp.Body.Close()
		return err
	}

	return nil
}

func handleAlivePost(rw http.ResponseWriter, request *http.Request) {
	aliverequest, err := parseAlivePost(request.Body)
	if err != nil {
		log.Printf("ERROR: Couldn't parse request -- %v", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("DeviceID: %s, Timeout: %d\n", aliverequest.DeviceID, aliverequest.Timeout)

	mutex.Lock()
	timer, timerFound := timersMap[aliverequest.DeviceID]
	mutex.Unlock()

	if timerFound {
		timer.DeviceTimer.Reset(time.Millisecond * time.Duration(aliverequest.Timeout))
		rw.WriteHeader(http.StatusOK)
	} else {
		timer := time.NewTimer(time.Millisecond * time.Duration(aliverequest.Timeout))
		deviceTimer := DeviceTimer{aliverequest.DeviceID, timer}

		mutex.Lock()
		timersMap[aliverequest.DeviceID] = deviceTimer
		mutex.Unlock()

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

	client = &Client{
		HTTPClient: &http.Client{},
	}

	if err := http.ListenAndServe(
		fmt.Sprintf(
			"%s:%d", *serviceHost, *servicePort),
		Handlers()); err != nil {
		log.Fatalf("%v", err)
	}
}
