package main

import (
	"data_processing"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

var feeds_service_url string
var http_client *http.Client

func init() {
	feeds_service_url = os.Getenv("FEEDS_SERVICE_URL")

	// https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	http_client = &http.Client{
		Timeout: time.Second * 20,
	}
}

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(_ interface{}) ([]data_processing.ProviderAtomFeed, error) {
	service := data_processing.NewProviderFeedsService(feeds_service_url, http_client)
	return service.GetProviderFeeds()
}
