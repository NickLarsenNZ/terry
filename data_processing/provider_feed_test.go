package data_processing

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	dynamock "github.com/gusaul/go-dynamock"
)

var test_table string
var test_item_key_value string
var test_item_record struct {
	etag        string
	lastversion string
}
var mock_dynamodb dynamodbiface.DynamoDBAPI
var mock *dynamock.DynaMock

func init() {
	mock_dynamodb, mock = dynamock.New()
	test_table = "example"

	test_item_key_value = "aws"
	test_item_record = struct {
		etag        string
		lastversion string
	}{etag: "33a64df551425fcc55e4d42a148795d9f25f89d4", lastversion: "v1.6.5"}

	mock.ExpectGetItem().ToTable(test_table).WithKeys(map[string]*dynamodb.AttributeValue{
		"Provider": {
			S: aws.String(test_item_key_value),
		},
	}).WillReturns(dynamodb.GetItemOutput{
		Item: map[string]*dynamodb.AttributeValue{
			"Provider": {S: aws.String(test_item_key_value)},
			"ETag":     {S: aws.String(test_item_record.etag)},
			"Version":  {S: aws.String(test_item_record.lastversion)},
		},
	})
}

func TestLastEtag(t *testing.T) {
	service := NewProviderFeedService(test_table, mock_dynamodb)

	// test where we do have etags
	if actual_etag, err := service.LastETag(test_item_key_value); actual_etag != test_item_record.etag {
		if err != nil {
			t.Fatalf("Unexpected errror: %v", err)
		}
		t.Fatalf("expected %s, got %s", test_item_record.etag, actual_etag)
	}

	// test for error when there is no etag found
	if _, err := service.LastETag("non_existant"); !errors.Is(err, ErrETagNotFound) {
		t.Fatalf("expected %v, got %v", ErrETagNotFound, err)
	}
}
