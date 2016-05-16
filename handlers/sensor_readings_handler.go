package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/skidder/streammarker-data-access/db"
)

// SensorReadingsHandler instance for retrieving readings
type SensorReadingsHandler struct {
	deviceManager        db.DeviceManager
	measurementsDatabase db.MeasurementsDatabase
}

// NewSensorReadingsHandler creates a new SensorReadingsHandler
func NewSensorReadingsHandler(deviceManager db.DeviceManager, measurementsDatabase db.MeasurementsDatabase) *SensorReadingsHandler {
	return &SensorReadingsHandler{deviceManager, measurementsDatabase}
}

// InitializeRouterForSensorsDataRetrieval creates a SensorReadingsHandler on the given router
func InitializeRouterForSensorsDataRetrieval(r *mux.Router, deviceManager db.DeviceManager, measurementsDatabase db.MeasurementsDatabase) {
	m := NewSensorReadingsHandler(deviceManager, measurementsDatabase)
	r.HandleFunc("/data-access/v1/sensors/account/{account_id}", m.GetSensors).Methods("GET")
	r.HandleFunc("/data-access/v1/last_sensor_readings/account/{account_id}", m.GetLastSensorReadings).Methods("GET")
	r.HandleFunc("/data-access/v1/sensor_readings", m.QueryForSensorReadings).Methods("GET")
}

// GetSensors retrieves a list of sensors in an account
func (m *SensorReadingsHandler) GetSensors(resp http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	state := q.Get("state")

	accountID := mux.Vars(req)["account_id"]
	if sensors, err := m.deviceManager.GetSensors(accountID, state); err == nil {
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusOK)
		responseEncoder := json.NewEncoder(resp)
		responseEncoder.Encode(&GetSensorsResponse{sensors})
	} else {
		log.Printf("Error getting sensors for account: %s", err.Error())
		http.Error(resp,
			"Error getting sensors for account",
			http.StatusInternalServerError)
	}
}

// GetLastSensorReadings retrieves last sensor readings for sensors in an account
func (m *SensorReadingsHandler) GetLastSensorReadings(resp http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	state := q.Get("state")

	accountID := mux.Vars(req)["account_id"]
	if sensors, err := m.measurementsDatabase.GetLastSensorReadings(accountID, state); err == nil {
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusOK)
		responseEncoder := json.NewEncoder(resp)
		responseEncoder.Encode(sensors)
	} else {
		log.Printf("Error getting last sensor readings for account: %s", err.Error())
		http.Error(resp,
			"Error getting last sensor readings for account",
			http.StatusInternalServerError)
	}
}

// QueryForSensorReadings retrieves readings for a sensor in an account matching certain criteria
func (m *SensorReadingsHandler) QueryForSensorReadings(resp http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	accountID := q.Get("account_id")
	sensorID := q.Get("sensor_id")
	var err error
	var startTime, endTime int64
	if q.Get("start_time") != "" {
		if startTime, err = strconv.ParseInt(q.Get("start_time"), 10, 32); err != nil {
			log.Printf("Unable to parse start_time as int: %s", err.Error())
			http.Error(resp, "Unable to parse start_time as int", http.StatusBadRequest)
			return
		}
	} else {
		// default start-time is one month ago
		startTime = time.Now().AddDate(0, -1, 0).Unix()
	}
	if q.Get("end_time") != "" {
		if endTime, err = strconv.ParseInt(q.Get("end_time"), 10, 32); err != nil {
			log.Printf("Unable to parse end_time as int: %s", err.Error())
			http.Error(resp, "Unable to parse end_time as int", http.StatusBadRequest)
			return
		}
	} else {
		endTime = time.Now().Unix()
	}
	if sensorReadings, err := m.measurementsDatabase.QueryForSensorReadings(accountID, sensorID, startTime, endTime); err == nil {
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusOK)
		responseEncoder := json.NewEncoder(resp)
		responseEncoder.Encode(sensorReadings)
	} else {
		log.Printf("Error querying for sensor readings for account: %s", err.Error())
		http.Error(resp,
			"Error querying for sensor readings for account",
			http.StatusInternalServerError)
	}
}

// GetSensorsResponse has a set of sensors
type GetSensorsResponse struct {
	Sensors []*db.Sensor `json:"sensors"`
}
