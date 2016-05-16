package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mholt/binding"
	"github.com/skidder/streammarker-data-access/db"
)

// SensorHandler instance
type SensorHandler struct {
	database db.DeviceManager
}

// NewSensorHandler creates a new SensorHandler
func NewSensorHandler(database db.DeviceManager) *SensorHandler {
	return &SensorHandler{database}
}

// InitializeRouterForSensorHandler initializes the handler on the given router
func InitializeRouterForSensorHandler(r *mux.Router, database db.DeviceManager) {
	m := NewSensorHandler(database)
	r.HandleFunc("/data-access/v1/sensor/{sensor_id}", m.GetSensor).Methods("GET")
	r.HandleFunc("/data-access/v1/sensor/{sensor_id}", m.UpdateSensor).Methods("PUT")
}

// GetSensor retrieves a sensor from the database
func (m *SensorHandler) GetSensor(resp http.ResponseWriter, req *http.Request) {
	sensorID := mux.Vars(req)["sensor_id"]
	if sensor, err := m.database.GetSensor(sensorID); err == nil {
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

// UpdateSensor updates a sensor record in the database
func (m *SensorHandler) UpdateSensor(resp http.ResponseWriter, req *http.Request) {
	// bind the request to a sensor model
	sensorUpdates := new(db.Sensor)
	errs := binding.Bind(req, sensorUpdates)
	if errs.Handle(resp) {
		log.Printf("Error while binding request to model: %s", errs.Error())
		http.Error(resp,
			"Invalid request",
			http.StatusBadRequest)
		return
	}

	sensorID := mux.Vars(req)["sensor_id"]
	if _, err := m.database.GetSensor(sensorID); err != nil {
		log.Printf("Error getting sensor for update: %s", err.Error())
		http.Error(resp,
			"Error getting sensor for update",
			http.StatusBadRequest)
		return
	}
	if sensor, err := m.database.UpdateSensor(sensorID, sensorUpdates); err == nil {
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
