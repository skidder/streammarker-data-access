package dao

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/mholt/binding"

	"github.com/urlgrey/streammarker-data-access/geo"
)

const (
	tableTimestampFormat = "2006-01"
)

var (
	dynamoPutAction = "PUT"
)

type Database struct {
	dynamoDBService *dynamodb.DynamoDB
	geoLookup       *geo.GoogleGeoLookup
}

func NewDatabase(dynamoDBService *dynamodb.DynamoDB, geoLookup *geo.GoogleGeoLookup) *Database {
	return &Database{dynamoDBService: dynamoDBService, geoLookup: geoLookup}
}

// GetTableWaitTime returns table wait time
func (d *Database) GetTableWaitTime() time.Duration {
	var waitTime string
	if waitTime = os.Getenv("STREAMMARKER_DYNAMO_WAIT_TIME"); waitTime == "" {
		waitTime = "30s"
	}

	t, err := time.ParseDuration(waitTime)
	if err != nil {
		return 30 * time.Second
	}
	return t
}

// GetAccount returns account record for given account ID
func (d *Database) GetAccount(accountID string) (*Account, error) {
	params := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(accountID),
			},
		},
		TableName: aws.String("accounts"),
		AttributesToGet: []*string{
			aws.String("name"),
			aws.String("state"),
		},
		ConsistentRead: aws.Bool(true),
	}

	resp, err := d.dynamoDBService.GetItem(params)
	if err == nil {
		if resp.Item != nil {
			account := &Account{
				ID:    accountID,
				Name:  *resp.Item["name"].S,
				State: *resp.Item["state"].S,
			}
			return account, nil
		}
		return nil, fmt.Errorf("Account not found: %s", accountID)
	}
	return nil, err
}

// GetRelay returns relay record for given ID
func (d *Database) GetRelay(relayID string) (*Relay, error) {
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
func (d *Database) UpdateSensor(sensorID string, sensorUpdates *Sensor) (*Sensor, error) {
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
func (d *Database) GetSensor(sensorID string) (*Sensor, error) {
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
func (d *Database) GetSensors(accountID string, state string) ([]*Sensor, error) {
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
	if resp, err := d.dynamoDBService.Query(params); err == nil {
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
	} else {
		return nil, err
	}
}

// Get the latest sensor readings for the given account
func (d *Database) GetLastSensorReadings(accountID string, state string) (*LatestSensorReadings, error) {
	var sensors []*Sensor
	var err error
	if sensors, err = d.GetSensors(accountID, state); err != nil {
		return nil, err
	}

	latestReadings := &LatestSensorReadings{make(map[string]*SensorReading)}
	currentTime := time.Now()
	sensorReadingsTableName := fmt.Sprintf("sensor_readings_%s", currentTime.Format(tableTimestampFormat))
	for _, sensor := range sensors {
		params := &dynamodb.QueryInput{
			TableName:        aws.String(sensorReadingsTableName),
			Select:           aws.String("ALL_ATTRIBUTES"),
			ScanIndexForward: aws.Bool(false),
			KeyConditions: map[string]*dynamodb.Condition{
				"id": {
					ComparisonOperator: aws.String("EQ"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{
							S: aws.String(fmt.Sprintf("%s:%s", sensor.AccountID, sensor.ID)),
						},
					},
				},
			},
			Limit: aws.Int64(1),
		}

		reading := &SensorReading{
			SensorID:  sensor.ID,
			AccountID: sensor.AccountID,
			Name:      sensor.Name,
			State:     sensor.State,
		}
		var resp *dynamodb.QueryOutput
		resp, _ = d.dynamoDBService.Query(params)
		for _, sensorRecord := range resp.Items {
			var measurements []Measurement
			if err = json.Unmarshal([]byte(*sensorRecord["measurements"].S), &measurements); err != nil {
				return nil, err
			}
			reading.Measurements = measurements
			timestamp, _ := strconv.ParseInt(*sensorRecord["timestamp"].N, 10, 32)
			reading.Timestamp = int32(timestamp)
		}
		latestReadings.Sensors[reading.SensorID] = reading
	}

	return latestReadings, err
}

func (d *Database) QueryForHourlySensorReadings(accountID, sensorID string, startTime, endTime int64) (*QueryForHourlySensorReadingsResults, error) {
	endTimeString := strconv.FormatInt(endTime, 10)
	var err error

	results := &QueryForHourlySensorReadingsResults{accountID, sensorID, make([]*HourlyMeasurements, 0)}
	endTimeTS := time.Unix(endTime, 0)
	for i := 0; i < 3; i++ {
		monthReadings := make([]*HourlyMeasurements, 0)
		startTimeTS := time.Unix(startTime, 0)
		startTimeTS = startTimeTS.AddDate(0, i, 0)
		if i > 0 {
			startTimeTS = time.Date(startTimeTS.Year(), startTimeTS.Month(), 1, 0, 0, 0, 0, startTimeTS.Location())
		}
		if endTimeTS.Before(startTimeTS) {
			break
		}

		startTimeString := strconv.FormatInt(startTimeTS.Unix(), 10)
		sensorReadingsTableName := fmt.Sprintf("hourly_sensor_readings_%s", startTimeTS.Format(tableTimestampFormat))
		params := &dynamodb.QueryInput{
			TableName:        aws.String(sensorReadingsTableName),
			Select:           aws.String("ALL_ATTRIBUTES"),
			ScanIndexForward: aws.Bool(false),
			Limit:            aws.Int64(10000),
			KeyConditions: map[string]*dynamodb.Condition{
				"id": {
					ComparisonOperator: aws.String("EQ"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{
							S: aws.String(fmt.Sprintf("%s:%s", accountID, sensorID)),
						},
					},
				},
				"timestamp": {
					ComparisonOperator: aws.String("BETWEEN"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{
							N: &startTimeString,
						},
						{
							N: &endTimeString,
						},
					},
				},
			},
		}

		var resp *dynamodb.QueryOutput
		if resp, err = d.dynamoDBService.Query(params); err == nil {
			for _, sensorRecord := range resp.Items {
				timestamp, _ := strconv.ParseInt(*sensorRecord["timestamp"].N, 10, 32)
				readings := &HourlyMeasurements{Timestamp: int32(timestamp)}
				if err = json.Unmarshal([]byte(*sensorRecord["measurements"].S), &readings.Measurements); err != nil {
					return nil, err
				}
				monthReadings = append(monthReadings, readings)
			}
		}
		results.Readings = append(monthReadings, results.Readings...)
	}
	return results, err
}

func (d *Database) QueryForSensorReadings(accountID, sensorID string, startTime, endTime int64) (*QueryForSensorReadingsResults, error) {
	endTimeString := strconv.FormatInt(endTime, 10)
	var err error

	results := &QueryForSensorReadingsResults{accountID, sensorID, make([]*MinimalReading, 0)}
	endTimeTS := time.Unix(endTime, 0)
	for i := 0; i < 3; i++ {
		monthReadings := make([]*MinimalReading, 0)
		startTimeTS := time.Unix(startTime, 0)
		startTimeTS = startTimeTS.AddDate(0, i, 0)
		if i > 0 {
			startTimeTS = time.Date(startTimeTS.Year(), startTimeTS.Month(), 1, 0, 0, 0, 0, startTimeTS.Location())
		}
		if endTimeTS.Before(startTimeTS) {
			break
		}

		startTimeString := strconv.FormatInt(startTimeTS.Unix(), 10)
		sensorReadingsTableName := fmt.Sprintf("sensor_readings_%s", startTimeTS.Format(tableTimestampFormat))
		params := &dynamodb.QueryInput{
			TableName:        aws.String(sensorReadingsTableName),
			Select:           aws.String("ALL_ATTRIBUTES"),
			ScanIndexForward: aws.Bool(false),
			Limit:            aws.Int64(10000),
			KeyConditions: map[string]*dynamodb.Condition{
				"id": {
					ComparisonOperator: aws.String("EQ"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{
							S: aws.String(fmt.Sprintf("%s:%s", accountID, sensorID)),
						},
					},
				},
				"timestamp": {
					ComparisonOperator: aws.String("BETWEEN"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{
							N: &startTimeString,
						},
						{
							N: &endTimeString,
						},
					},
				},
			},
		}

		var resp *dynamodb.QueryOutput
		if resp, err = d.dynamoDBService.Query(params); err == nil {
			for _, sensorRecord := range resp.Items {
				var measurements []Measurement
				if err = json.Unmarshal([]byte(*sensorRecord["measurements"].S), &measurements); err != nil {
					return nil, err
				}
				timestamp, _ := strconv.ParseInt(*sensorRecord["timestamp"].N, 10, 32)
				minimalReading := &MinimalReading{int32(timestamp), measurements}
				monthReadings = append(monthReadings, minimalReading)
			}
		}
		err = nil
		results.Readings = append(monthReadings, results.Readings...)
	}
	return results, err
}

type Relay struct {
	ID        string `json:"id"`
	AccountID string `json:"account_id"`
	Name      string `json:"name"`
	State     string `json:"state"`
}

type QueryForSensorReadingsResults struct {
	AccountID string            `json:"account_id"`
	SensorID  string            `json:"sensor_id"`
	Readings  []*MinimalReading `json:"readings"`
}

type MinimalReading struct {
	Timestamp    int32         `json:"timestamp"`
	Measurements []Measurement `json:"measurements"`
}

type LatestSensorReadings struct {
	Sensors map[string]*SensorReading `json:"sensors"`
}

type SensorReading struct {
	SensorID     string        `json:"sensor_id"`
	AccountID    string        `json:"account_id"`
	Name         string        `json:"name"`
	State        string        `json:"state"`
	Timestamp    int32         `json:"timestamp"`
	Measurements []Measurement `json:"measurements"`
}

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

type Account struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

type Measurement struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type MinMaxMeasurement struct {
	Name string      `json:"name"`
	Min  Measurement `json:"min"`
	Max  Measurement `json:"max"`
}

type HourlyMeasurements struct {
	Timestamp    int32                `json:"timestamp"`
	Measurements []*MinMaxMeasurement `json:"measurements"`
}

type QueryForHourlySensorReadingsResults struct {
	AccountID string                `json:"account_id"`
	SensorID  string                `json:"sensor_id"`
	Readings  []*HourlyMeasurements `json:"readings"`
}

func (s *Sensor) FieldMap() binding.FieldMap {
	return binding.FieldMap{
		&s.Name:            "name",
		&s.State:           "state",
		&s.LocationEnabled: "location_enabled",
		&s.Latitude:        "latitude",
		&s.Longitude:       "longitude",
		&s.SampleFrequency: "sample_frequency",
	}
}
