package main

import (
	"data_processing"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type X struct {
	Provider  string
	Version   string
	PluginUrl string
}

var dynamodb_table string
var dynamodb_client dynamodbiface.DynamoDBAPI
var http_client *http.Client

func init() {
	dynamodb_table = os.Getenv("DYNAMODB_TABLE")
	region := os.Getenv("AWS_REGION")

	mySession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region)}),
	)

	// Create a DynamoDB client from a session.
	dynamodb_client = dynamodb.New(mySession)

	// https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	http_client = &http.Client{
		Timeout: time.Second * 20,
	}
}

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(e data_processing.ProviderAtomFeed) ([]data_processing.ProviderVersion, error) {
	service := data_processing.NewProviderFeedService(dynamodb_table, dynamodb_client, http_client)
	return service.CheckForNewVersions(e)
}
