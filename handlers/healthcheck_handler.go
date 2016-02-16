package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gorilla/mux"
)

type HealthCheckHandler struct {
	dynamoDB *dynamodb.DynamoDB
}

func NewHealthCheckHandler(dynamoDB *dynamodb.DynamoDB) *HealthCheckHandler {
	return &HealthCheckHandler{dynamoDB}
}

// Add routes to router
func InitializeRouterForHealthCheckHandler(r *mux.Router, dynamoDB *dynamodb.DynamoDB) {
	m := NewHealthCheckHandler(dynamoDB)
	r.HandleFunc("/healthcheck", m.HealthCheck).Methods("GET")
}

// Examine and report the health of the component and dependencies
func (h *HealthCheckHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	params := &dynamodb.DescribeTableInput{
		TableName: aws.String("sensors"), // Required
	}
	var err error
	if _, err = h.dynamoDB.DescribeTable(params); err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			// Generic AWS Error with Code, Message, and original error (if any)
			log.Printf("AWS error code=%s, Message=%s, Original error=%s", awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				// A service error occurred
				log.Printf("Request error code=%s, Message=%s, Original error=%s", reqErr.Code(), reqErr.Message(), reqErr.OrigErr())
			}
		} else {
			// This case should never be hit, The SDK should alwsy return an
			// error which satisfies the awserr.Error interface.
			log.Printf("Generic error: %s", err.Error())
		}
		http.Error(w,
			fmt.Sprintf("{\"error\": \"Error checking DynamoDB connectivity: %+v\"}", err),
			http.StatusInternalServerError)
	}
}
