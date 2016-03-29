package db

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/influxdata/influxdb/models"
)

const (
	sensorMeasurementsTableName = "sensor_measurements"
)

// MeasurementsDatabase provides functions for retrieving sensor measurements
type MeasurementsDatabase interface {
	GetLastSensorReadings(string, string) (*LatestSensorReadings, error)
	QueryForSensorReadings(string, string, int64, int64) (*QueryForSensorReadingsResults, error)
}

// InfluxDAO represents a DAO capable of interacting with InfluxDB
type InfluxDAO struct {
	c             client.Client
	databaseName  string
	deviceManager DeviceManager
}

// NewInfluxDAO creates a new DAO for interacting with InfluxDB
func NewInfluxDAO(address string, username string, password string, databaseName string, deviceManager DeviceManager) (*InfluxDAO, error) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     address,
		Username: username,
		Password: password,
	})
	return &InfluxDAO{c, databaseName, deviceManager}, err
}

// GetLastSensorReadings returns the latest sensor readings for the given account
func (i *InfluxDAO) GetLastSensorReadings(accountID string, state string) (*LatestSensorReadings, error) {
	var sensors []*Sensor
	var err error
	if sensors, err = i.deviceManager.GetSensors(accountID, state); err != nil {
		return nil, err
	}

	latestReadings := &LatestSensorReadings{make(map[string]*SensorReading)}
	for _, sensor := range sensors {
		reading := &SensorReading{
			SensorID:  sensor.ID,
			AccountID: sensor.AccountID,
			Name:      sensor.Name,
			State:     sensor.State,
		}

		series, err := i.getLastReadingForSensor(sensor.ID, sensor.AccountID)
		if err != nil {
			return latestReadings, err
		}

		if series != nil {
			reading.Measurements = make([]Measurement, 0)
			for key, value := range series.Columns {
				if strings.Contains(value, "time") {
					var timestamp time.Time
					timestamp, err = time.Parse(time.RFC3339, series.Values[0][key].(string))
					if err != nil {
						return latestReadings, err
					}
					reading.Timestamp = timestamp.Unix()
				} else if strings.Contains(value, "temperature") {
					var temperatureValue float64
					temperatureValue, err = series.Values[0][key].(json.Number).Float64()
					if err == nil {
						reading.Measurements = append(reading.Measurements, Measurement{
							Name:  value,
							Unit:  "Celsius",
							Value: temperatureValue,
						})
					}
				} else if strings.Contains(value, "humidity") {
					var humidityValue float64
					humidityValue, err = series.Values[0][key].(json.Number).Float64()
					if err == nil {
						reading.Measurements = append(reading.Measurements, Measurement{
							Name:  value,
							Unit:  "%",
							Value: humidityValue,
						})
					}
				} else if strings.Contains(value, "soil_moisture") {
					var soilMoistureValue float64
					soilMoistureValue, err = series.Values[0][key].(json.Number).Float64()
					if err == nil {
						reading.Measurements = append(reading.Measurements, Measurement{
							Name:  value,
							Unit:  "VWC",
							Value: soilMoistureValue,
						})
					}
				}
			}
		}
		latestReadings.Sensors[reading.SensorID] = reading
	}

	return latestReadings, err
}

// QueryForSensorReadings returns sensor readings within an account
func (i *InfluxDAO) QueryForSensorReadings(accountID, sensorID string, startTime, endTime int64) (*QueryForSensorReadingsResults, error) {
	results := &QueryForSensorReadingsResults{accountID, sensorID, make([]*MinimalReading, 0)}

	res, err := i.queryDB(fmt.Sprintf("SELECT * from %s where sensor_id = '%s' and account_id = '%s' and time >= '%s' and time <= '%s' order by time desc", sensorMeasurementsTableName, sensorID, accountID, time.Unix(startTime, 0).Format(time.RFC3339), time.Unix(endTime, 0).Format(time.RFC3339)))
	if err != nil {
		return nil, err
	}
	if len(res) != 1 || res[0].Series == nil || len(res[0].Series) == 0 {
		return results, nil
	}

	row := res[0].Series[0]

	for _, rowValues := range row.Values {
		rowReading := &MinimalReading{
			Measurements: make([]Measurement, 0),
		}
		for k, v := range rowValues {
			valueName := row.Columns[k]
			if strings.Contains(valueName, "time") {
				var timestamp time.Time
				timestamp, err = time.Parse(time.RFC3339, v.(string))
				if err != nil {
					return results, err
				}
				rowReading.Timestamp = timestamp.Unix()
			} else if strings.Contains(valueName, "temperature") {
				var temperatureValue float64
				temperatureValue, err = v.(json.Number).Float64()
				if err == nil {
					rowReading.Measurements = append(rowReading.Measurements, Measurement{
						Name:  valueName,
						Unit:  "Celsius",
						Value: temperatureValue,
					})
				}
			} else if strings.Contains(valueName, "humidity") {
				var humidityValue float64
				humidityValue, err = v.(json.Number).Float64()
				if err == nil {
					rowReading.Measurements = append(rowReading.Measurements, Measurement{
						Name:  valueName,
						Unit:  "%",
						Value: humidityValue,
					})
				}
			} else if strings.Contains(valueName, "soil_moisture") {
				var soilMoistureValue float64
				soilMoistureValue, err = v.(json.Number).Float64()
				if err == nil {
					rowReading.Measurements = append(rowReading.Measurements, Measurement{
						Name:  valueName,
						Unit:  "VWC",
						Value: soilMoistureValue,
					})
				}
			}
		}
		results.Readings = append(results.Readings, rowReading)
	}
	return results, err
}

func (i *InfluxDAO) getLastReadingForSensor(sensorID string, accountID string) (*models.Row, error) {
	res, err := i.queryDB(fmt.Sprintf("SELECT * from %s where sensor_id = '%s' and account_id = '%s' order by time desc limit 1", sensorMeasurementsTableName, sensorID, accountID))
	if err != nil {
		return nil, err
	}
	if len(res) != 1 || res[0].Series == nil {
		// the query returned no rows, must be empty
		return nil, nil
	}
	return &res[0].Series[0], err
}

// queryDB convenience function to query the database
func (i *InfluxDAO) queryDB(cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: i.databaseName,
	}
	if response, err := i.c.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}
