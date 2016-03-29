package main // import "github.com/urlgrey/streammarker-data-access"

import (
	"fmt"
	"os"

	"github.com/urlgrey/streammarker-data-access/db"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/urlgrey/streammarker-data-access/geo"
	"github.com/urlgrey/streammarker-data-access/handlers"
)

const (
	defaultInfluxDBUsername = "streammarker"
	defaultInfluxDBAddress  = "http://127.0.0.1:8086"
	defaultInfluxDBName     = "streammarker_measurements"
)

func main() {
	mainServer := negroni.New()

	// Token auth middleware
	tokenVerification := handlers.NewTokenVerificationMiddleware()
	tokenVerification.Initialize()
	mainServer.Use(negroni.NewRecovery())
	mainServer.Use(negroni.NewLogger())
	mainServer.Use(negroni.HandlerFunc(tokenVerification.Run))
	mainServer.Use(gzip.Gzip(gzip.DefaultCompression))

	// Create external service connections
	s := session.New()
	dynamoDBConnection := createDynamoDBConnection(s)
	geoLookup := geo.NewGoogleGeoLookup(os.Getenv("GOOGLE_API_KEY"))
	geoLookup.Initialize()
	deviceDatabase := db.NewDeviceDatabase(dynamoDBConnection, geoLookup)
	measurementsDatabase, err := createMeasurementsDatabaseConnection(deviceDatabase)
	if err != nil {
		fmt.Printf("Error connecting to InfluxDB: %s\n", err.Error())
		return
	}

	// Initialize HTTP service handlers
	router := mux.NewRouter()
	handlers.InitializeRouterForSensorsDataRetrieval(router, deviceDatabase, measurementsDatabase)
	handlers.InitializeRouterForSensorHandler(router, deviceDatabase)
	mainServer.UseHandler(router)
	go mainServer.Run(":3000")

	// Run healthcheck service
	healthCheckServer := negroni.New()
	healthCheckRouter := mux.NewRouter()
	handlers.InitializeRouterForHealthCheckHandler(healthCheckRouter, dynamoDBConnection)
	healthCheckServer.UseHandler(healthCheckRouter)
	healthCheckServer.Run(":3100")
}

func createDynamoDBConnection(s *session.Session) *dynamodb.DynamoDB {
	config := &aws.Config{}
	if endpoint := os.Getenv("STREAMMARKER_DYNAMO_ENDPOINT"); endpoint != "" {
		config.Endpoint = &endpoint
	}

	return dynamodb.New(s, config)
}

func createMeasurementsDatabaseConnection(deviceManager db.DeviceManager) (db.MeasurementsDatabase, error) {
	influxDBUsername := os.Getenv("STREAMMARKER_INFLUXDB_USERNAME")
	if influxDBUsername == "" {
		influxDBUsername = defaultInfluxDBUsername
	}
	influxDBPassword := os.Getenv("STREAMMARKER_INFLUXDB_PASSWORD")
	influxDBAddress := os.Getenv("STREAMMARKER_INFLUXDB_ADDRESS")
	if influxDBAddress == "" {
		influxDBAddress = defaultInfluxDBAddress
	}

	influxDBName := os.Getenv("STREAMMARKER_INFLUXDB_NAME")
	if influxDBName == "" {
		influxDBName = defaultInfluxDBName
	}
	return db.NewInfluxDAO(influxDBAddress, influxDBUsername, influxDBPassword, influxDBName, deviceManager)
}
