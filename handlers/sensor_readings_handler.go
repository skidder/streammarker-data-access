package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/urlgrey/streammarker-data-access/dao"
)

type SensorReadingsHandler struct {
	database *dao.Database
}

func NewSensorReadingsHandler(database *dao.Database) *SensorReadingsHandler {
	return &SensorReadingsHandler{database}
}

func InitializeRouterForSensorsDataRetrieval(r *mux.Router, database *dao.Database) {
	m := NewSensorReadingsHandler(database)
	r.HandleFunc("/data-access/v1/sensors/account/{account_id}", m.GetSensors).Methods("GET")
	r.HandleFunc("/data-access/v1/last_sensor_readings/account/{account_id}", m.GetLastSensorReadings).Methods("GET")
	r.HandleFunc("/data-access/v1/sensor_readings", m.QueryForSensorReadings).Methods("GET")
	r.HandleFunc("/data-access/v1/hourly_sensor_readings", m.QueryForHourlySensorReadings).Methods("GET")
}

func (m *SensorReadingsHandler) GetSensors(resp http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	state := q.Get("state")

	accountId := mux.Vars(req)["account_id"]
	if sensors, err := m.database.GetSensors(accountId, state); err == nil {
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

func (m *SensorReadingsHandler) GetLastSensorReadings(resp http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	state := q.Get("state")

	accountId := mux.Vars(req)["account_id"]
	if sensors, err := m.database.GetLastSensorReadings(accountId, state); err == nil {
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

func (m *SensorReadingsHandler) QueryForSensorReadings(resp http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	accountId := q.Get("account_id")
	sensorId := q.Get("sensor_id")
	var err error
	var startTime, endTime int64
	if q.Get("start_time") != "" {
		if startTime, err = strconv.ParseInt(q.Get("start_time"), 10, 32); err != nil {
			log.Printf("Unable to parse start_time as int", err.Error())
			http.Error(resp, "Unable to parse start_time as int", http.StatusBadRequest)
			return
		}
	} else {
		// default start-time is one month ago
		startTime = time.Now().AddDate(0, -1, 0).Unix()
	}
	if q.Get("end_time") != "" {
		if endTime, err = strconv.ParseInt(q.Get("end_time"), 10, 32); err != nil {
			log.Printf("Unable to parse end_time as int", err.Error())
			http.Error(resp, "Unable to parse end_time as int", http.StatusBadRequest)
			return
		}
	} else {
		endTime = time.Now().Unix()
	}
	if sensorReadings, err := m.database.QueryForSensorReadings(accountId, sensorId, startTime, endTime); err == nil {
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

func (m *SensorReadingsHandler) QueryForHourlySensorReadings(resp http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	accountId := q.Get("account_id")
	sensorId := q.Get("sensor_id")
	var err error
	var startTime, endTime int64
	if q.Get("start_time") != "" {
		if startTime, err = strconv.ParseInt(q.Get("start_time"), 10, 32); err != nil {
			log.Printf("Unable to parse start_time as int", err.Error())
			http.Error(resp, "Unable to parse start_time as int", http.StatusBadRequest)
		}
	} else {
		// default start-time is one month ago
		startTime = time.Now().AddDate(0, -1, 0).Unix()
	}
	if q.Get("end_time") != "" {
		if endTime, err = strconv.ParseInt(q.Get("end_time"), 10, 32); err != nil {
			log.Printf("Unable to parse end_time as int", err.Error())
			http.Error(resp, "Unable to parse end_time as int", http.StatusBadRequest)
		}
	} else {
		endTime = time.Now().Unix()
	}
	if sensorReadings, err := m.database.QueryForHourlySensorReadings(accountId, sensorId, startTime, endTime); err == nil {
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusOK)
		responseEncoder := json.NewEncoder(resp)
		responseEncoder.Encode(sensorReadings)
	} else {
		log.Printf("Error querying for hourly sensor readings for account: %s", err.Error())
		http.Error(resp,
			"Error querying for sensor readings for account",
			http.StatusInternalServerError)
	}
}

type GetSensorsResponse struct {
	Sensors []*dao.Sensor `json:"sensors"`
}

type GetLastSensorReadingsResponse struct {
	Sensors []*dao.Sensor `json:"sensors"`
}
