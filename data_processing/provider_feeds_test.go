package data_processing

import (
	"testing"

	mockhttp "github.com/nicklarsennz/mockhttp"
	"github.com/pkg/errors"
)

func TestGetFeeds(t *testing.T) {
	// Instantiate a new http.Client
	client, err := mockhttp.NewClient("./responders.yml")
	if err != nil {
		t.Errorf(errors.Wrap(err, "mockhttp.NewClient()").Error())
	}

	service := NewProviderFeedsService("http://localhost:8080/feeds", client)

	// Inject the mock client into the real app
	feeds, err := service.GetProviderFeeds()
	if err != nil {
		t.Errorf(errors.Wrap(err, "GetProviderFeeds()").Error())
	}

	expected := 2 // Number of feeds returned, as per responders.yaml
	actual := len(feeds)
	if actual != expected {
		t.Errorf("expected %d, got %d", expected, actual)
	}
}
