package main // import "github.com/urlgrey/streammarker-data-access"

import (
	"os"

	"github.com/urlgrey/streammarker-data-access/dao"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/urlgrey/streammarker-data-access/geo"
	"github.com/urlgrey/streammarker-data-access/handlers"
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
	db := dao.NewDatabase(dynamoDBConnection, geoLookup)

	// Initialize HTTP service handlers
	router := mux.NewRouter()
	handlers.InitializeRouterForSensorsDataRetrieval(router, db)
	handlers.InitializeRouterForSensorHandler(router, db)
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
