package db

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/mholt/binding"

	"github.com/skidder/streammarker-data-access/geo"
)

const (
	tableTimestampFormat = "2006-01"
)

var (
	dynamoPutAction = "PUT"
)

// Database can be used to read and write sensor & relay data
type deviceDatabase struct {
	dynamoDBService *dynamodb.DynamoDB
	geoLookup       *geo.GoogleGeoLookup
}

// DeviceManager provides functions for updating and retrieving sensor and relay devices
type DeviceManager interface {
	GetRelay(string) (*Relay, error)
	GetSensor(string) (*Sensor, error)
	GetSensors(string, string) ([]*Sensor, error)
	UpdateSensor(string, *Sensor) (*Sensor, error)
}

// NewDeviceDatabase constructs a new Database instance
func NewDeviceDatabase(dynamoDBService *dynamodb.DynamoDB, geoLookup *geo.GoogleGeoLookup) DeviceManager {
	return &deviceDatabase{dynamoDBService: dynamoDBService, geoLookup: geoLookup}
}

// GetRelay returns relay record for given ID
func (d *deviceDatabase) GetRelay(relayID string) (*Relay, error) {
	params := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(relayID),
			},
		},
		TableName: aws.String("relays"),
		AttributesToGet: []*string{
			aws.String("account_id"),
			aws.String("name"),
			aws.String("state"),
		},
		ConsistentRead: aws.Bool(true),
	}

	resp, err := d.dynamoDBService.GetItem(params)
	if err == nil {
		if resp.Item != nil {
			relay := &Relay{
				ID:        relayID,
				AccountID: *resp.Item["account_id"].S,
				Name:      *resp.Item["name"].S,
				State:     *resp.Item["state"].S,
			}
			return relay, nil
		}
		return nil, fmt.Errorf("Relay not found: %s", relayID)
	}
	return nil, err
}

// UpdateSensor updates sensor database record
func (d *deviceDatabase) UpdateSensor(sensorID string, sensorUpdates *Sensor) (*Sensor, error) {
	params := &dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(sensorID),
			},
		},
		TableName: aws.String("sensors"),
		AttributeUpdates: map[string]*dynamodb.AttributeValueUpdate{
			"name": {
				Action: &dynamoPutAction,
				Value: &dynamodb.AttributeValue{
					S: aws.String(sensorUpdates.Name),
				},
			},
			"state": {
				Action: &dynamoPutAction,
				Value: &dynamodb.AttributeValue{
					S: aws.String(sensorUpdates.State),
				},
			},
			"location_enabled": {
				Action: &dynamoPutAction,
				Value: &dynamodb.AttributeValue{
					BOOL: aws.Bool(sensorUpdates.LocationEnabled),
				},
			},
			"latitude": {
				Action: &dynamoPutAction,
				Value: &dynamodb.AttributeValue{
					N: aws.String(fmt.Sprintf("%f", sensorUpdates.Latitude)),
				},
			},
			"longitude": {
				Action: &dynamoPutAction,
				Value: &dynamodb.AttributeValue{
					N: aws.String(fmt.Sprintf("%f", sensorUpdates.Longitude)),
				},
			},
			"sample_frequency": {
				Action: &dynamoPutAction,
				Value: &dynamodb.AttributeValue{
					N: aws.String(fmt.Sprintf("%d", sensorUpdates.SampleFrequency)),
				},
			},
		},
	}

	_, err := d.dynamoDBService.UpdateItem(params)
	if err == nil {
		return d.GetSensor(sensorID)
	}
	return nil, err
}

// GetSensor returns sensor record for the given sensor ID
func (d *deviceDatabase) GetSensor(sensorID string) (*Sensor, error) {
	params := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(sensorID),
			},
		},
		TableName: aws.String("sensors"),
		AttributesToGet: []*string{
			aws.String("account_id"),
			aws.String("name"),
			aws.String("state"),
			aws.String("location_enabled"),
			aws.String("latitude"),
			aws.String("longitude"),
			aws.String("sample_frequency"),
		},
		ConsistentRead: aws.Bool(true),
	}

	resp, err := d.dynamoDBService.GetItem(params)
	if err == nil {
		if resp.Item != nil {
			sensor := &Sensor{
				ID:              sensorID,
				AccountID:       *resp.Item["account_id"].S,
				Name:            *resp.Item["name"].S,
				State:           *resp.Item["state"].S,
				LocationEnabled: *resp.Item["location_enabled"].BOOL,
			}
			if resp.Item["sample_frequency"] != nil {
				sensor.SampleFrequency, _ = strconv.ParseInt(*resp.Item["sample_frequency"].N, 10, 64)
			} else {
				sensor.SampleFrequency = 1
			}

			if resp.Item["latitude"] != nil && resp.Item["longitude"] != nil {
				sensor.Latitude, _ = strconv.ParseFloat(*resp.Item["latitude"].N, 64)
				sensor.Longitude, _ = strconv.ParseFloat(*resp.Item["longitude"].N, 64)

				tz, err := d.geoLookup.FindTimezoneForLocation(sensor.Latitude, sensor.Longitude)
				if err == nil {
					sensor.TimeZoneID = tz.TimeZoneID
					sensor.TimeZoneName = tz.TimeZoneName
				}
			}
			return sensor, nil
		}
		return nil, fmt.Errorf("Sensor not found: %s", sensorID)
	}
	return nil, err
}

// GetSensors returns sensors for an account in a given state
func (d *deviceDatabase) GetSensors(accountID string, state string) ([]*Sensor, error) {
	params := &dynamodb.QueryInput{
		TableName: aws.String("sensors"),
		Select:    aws.String("ALL_PROJECTED_ATTRIBUTES"),
		KeyConditions: map[string]*dynamodb.Condition{
			"account_id": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(accountID),
					},
				},
			},
		},
		IndexName: aws.String("account_id-index"),
		Limit:     aws.Int64(100),
	}

	var sensors []*Sensor
	resp, err := d.dynamoDBService.Query(params)
	if err == nil {
		for _, sensorRecord := range resp.Items {
			if state != "" && *sensorRecord["state"].S != state {
				continue
			}

			s := &Sensor{
				ID:              *sensorRecord["id"].S,
				AccountID:       accountID,
				Name:            *sensorRecord["name"].S,
				State:           *sensorRecord["state"].S,
				LocationEnabled: *sensorRecord["location_enabled"].BOOL,
			}
			if sensorRecord["sample_frequency"] != nil {
				s.SampleFrequency, _ = strconv.ParseInt(*sensorRecord["sample_frequency"].N, 10, 64)
			} else {
				s.SampleFrequency = 1
			}

			if sensorRecord["latitude"] != nil && sensorRecord["longitude"] != nil {
				s.Latitude, _ = strconv.ParseFloat(*sensorRecord["latitude"].N, 64)
				s.Longitude, _ = strconv.ParseFloat(*sensorRecord["longitude"].N, 64)
			}
			sensors = append(sensors, s)
		}
		return sensors, nil
	}
	return nil, err
}

// Relay has details for a StreamMarker relay
type Relay struct {
	ID        string `json:"id"`
	AccountID string `json:"account_id"`
	Name      string `json:"name"`
	State     string `json:"state"`
}

// QueryForSensorReadingsResults has results for a sensor-readings query
type QueryForSensorReadingsResults struct {
	AccountID string            `json:"account_id"`
	SensorID  string            `json:"sensor_id"`
	Readings  []*MinimalReading `json:"readings"`
}

// MinimalReading represents a single reading
type MinimalReading struct {
	Timestamp    int64         `json:"timestamp"`
	Measurements []Measurement `json:"measurements"`
}

// LatestSensorReadings contains the latest readings for a set of sensors
type LatestSensorReadings struct {
	Sensors map[string]*SensorReading `json:"sensors"`
}

// SensorReading has details for a single reading
type SensorReading struct {
	SensorID     string        `json:"sensor_id"`
	AccountID    string        `json:"account_id"`
	Name         string        `json:"name"`
	State        string        `json:"state"`
	Timestamp    int64         `json:"timestamp"`
	Measurements []Measurement `json:"measurements"`
}

// Sensor represents a sensor capable of producing measurements
type Sensor struct {
	ID              string  `json:"id"`
	AccountID       string  `json:"account_id"`
	Name            string  `json:"name"`
	State           string  `json:"state"`
	LocationEnabled bool    `json:"location_enabled"`
	Latitude        float64 `json:"latitude,omitempty"`
	Longitude       float64 `json:"longitude,omitempty"`
	TimeZoneID      string  `json:"timezone_id,omitempty"`
	TimeZoneName    string  `json:"timezone_name,omitempty"`
	SampleFrequency int64   `json:"sample_frequency,omitempty"`
}

// Account has account details
type Account struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

// Measurement represents a single measurement's details
type Measurement struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

// MinMaxMeasurement has the minimum and maximum values for a measurement
type MinMaxMeasurement struct {
	Name string      `json:"name"`
	Min  Measurement `json:"min"`
	Max  Measurement `json:"max"`
}

// FieldMap binds Sensor value for JSON mapping
func (s *Sensor) FieldMap(req *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&s.Name:            "name",
		&s.State:           "state",
		&s.LocationEnabled: "location_enabled",
		&s.Latitude:        "latitude",
		&s.Longitude:       "longitude",
		&s.SampleFrequency: "sample_frequency",
	}
}
