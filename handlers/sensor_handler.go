package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mholt/binding"
	"github.com/urlgrey/streammarker-data-access/dao"
)

type SensorHandler struct {
	database *dao.Database
}

func NewSensorHandler(database *dao.Database) *SensorHandler {
	return &SensorHandler{database}
}

func InitializeRouterForSensorHandler(r *mux.Router, database *dao.Database) {
	m := NewSensorHandler(database)
	r.HandleFunc("/data-access/v1/sensor/{sensor_id}", m.GetSensor).Methods("GET")
	r.HandleFunc("/data-access/v1/sensor/{sensor_id}", m.UpdateSensor).Methods("PUT")
}

func (m *SensorHandler) GetSensor(resp http.ResponseWriter, req *http.Request) {
	sensorId := mux.Vars(req)["sensor_id"]
	if sensor, err := m.database.GetSensor(sensorId); err == nil {
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusOK)
		responseEncoder := json.NewEncoder(resp)
		responseEncoder.Encode(sensor)
	} else {
		log.Printf("Error getting sensor: %s", err.Error())
		http.Error(resp,
			"Error getting sensor",
			http.StatusInternalServerError)
	}
}

func (m *SensorHandler) UpdateSensor(resp http.ResponseWriter, req *http.Request) {
	// bind the request to a sensor model
	sensorUpdates := new(dao.Sensor)
	errs := binding.Bind(req, sensorUpdates)
	if errs.Handle(resp) {
		log.Printf("Error while binding request to model: %s", errs.Error())
		http.Error(resp,
			"Invalid request",
			http.StatusBadRequest)
		return
	}

	sensorId := mux.Vars(req)["sensor_id"]
	if _, err := m.database.GetSensor(sensorId); err != nil {
		log.Printf("Error getting sensor for update: %s", err.Error())
		http.Error(resp,
			"Error getting sensor for update",
			http.StatusBadRequest)
		return
	}
	if sensor, err := m.database.UpdateSensor(sensorId, sensorUpdates); err == nil {
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusOK)
		responseEncoder := json.NewEncoder(resp)
		responseEncoder.Encode(sensor)
	} else {
		log.Printf("Error getting sensor: %s", err.Error())
		http.Error(resp,
			"Error getting sensor",
			http.StatusInternalServerError)
	}
}
