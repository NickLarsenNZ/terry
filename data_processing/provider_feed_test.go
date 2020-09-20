package data_processing

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	dynamock "github.com/gusaul/go-dynamock"
)

var test_table = "example"
var test_item_key_value = "aws"
var test_item_record = struct {
	etag    string
	version string
}{etag: "33a64df551425fcc55e4d42a148795d9f25f89d4", version: "v1.6.5"}
var mock_dynamodb dynamodbiface.DynamoDBAPI
var mock *dynamock.DynaMock
var expectedKey map[string]*dynamodb.AttributeValue
var expectedReturn dynamodb.GetItemOutput

func init() {
	mock_dynamodb, mock = dynamock.New()

	expectedKey = map[string]*dynamodb.AttributeValue{
		"Provider": {
			S: aws.String(test_item_key_value),
		},
	}

	expectedReturn = dynamodb.GetItemOutput{
		Item: map[string]*dynamodb.AttributeValue{
			"Provider": {S: aws.String(test_item_key_value)},
			"ETag":     {S: aws.String(test_item_record.etag)},
			"Version":  {S: aws.String(test_item_record.version)},
		},
	}
}

func TestLastEtag(t *testing.T) {
	mock.ExpectGetItem().ToTable(test_table).WithKeys(expectedKey).WillReturns(expectedReturn)

	service := NewProviderFeedService(test_table, mock_dynamodb)

	// test where we do have etags
	if actual_etag, err := service.LastETag(test_item_key_value); actual_etag != test_item_record.etag {
		if err != nil {
			t.Fatalf("Unexpected errror: %v", err)
		}
		t.Fatalf("expected %s, got %s", test_item_record.etag, actual_etag)
	}

	// test for error when there is no etag found
	if _, err := service.LastETag("non_existant"); !errors.Is(err, ErrItemNotFound) {
		t.Fatalf("expected %v, got %v", ErrItemNotFound, err)
	}
}

func TestLastVersion(t *testing.T) {
	mock.ExpectGetItem().ToTable(test_table).WithKeys(expectedKey).WillReturns(expectedReturn)

	service := NewProviderFeedService(test_table, mock_dynamodb)

	// test where we do have versions
	if actual_version, err := service.LastVersion(test_item_key_value); actual_version != test_item_record.version {
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		t.Fatalf("expected %s, got %s", test_item_record.version, actual_version)
	}

	// test for error when there is no version found
	if _, err := service.LastVersion("non_existant"); !errors.Is(err, ErrItemNotFound) {
		t.Fatalf("expected %v, got %v", ErrItemNotFound, err)
	}
}

func TestCache(t *testing.T) {
	mock.ExpectGetItem().ToTable(test_table).WithKeys(expectedKey).WillReturns(expectedReturn)

	service := NewProviderFeedService(test_table, mock_dynamodb)

	// test first call which hits dynamodb
	_, err := service.LastVersion(test_item_key_value)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// subsequent call for the same key ,which should be returned from cache
	_, err = service.LastVersion(test_item_key_value)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Internal detail, but only way to ensure the cache was hit
	if service.cache_miss != 1 {
		t.Fatal("cache was not used")
	}
}
