package data_processing

import (
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	dynamock "github.com/gusaul/go-dynamock"
	mockhttp "github.com/nicklarsennz/mockhttp"
	"github.com/pkg/errors"
)

var test_table = "example"
var test_item_key_value = "aws"
var test_item_record = struct {
	etag    string
	version string
}{etag: "33a64df551425fcc55e4d42a148795d9f25f89d4", version: "v1.6.5"}
var test_atom_url = "https://github.com/terraform-providers/terraform-provider-aws/releases.atom"
var mock_dynamodb dynamodbiface.DynamoDBAPI
var mock *dynamock.DynaMock
var expectedKey map[string]*dynamodb.AttributeValue
var expectedReturn dynamodb.GetItemOutput
var mockHttpClient *http.Client

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

	// Instantiate a new http.Client
	mockHttpClient, _ = mockhttp.NewClient("./responders.yml")
	// if err != nil {
	// 	panic(errors.Wrap(err, "mockhttp.NewClient()").Error())
	// }
}

func TestLastEtag(t *testing.T) {
	mock.ExpectGetItem().ToTable(test_table).WithKeys(expectedKey).WillReturns(expectedReturn)
	mock.ExpectPutItem().ToTable(test_table)

	service := NewProviderFeedService(test_table, mock_dynamodb, mockHttpClient)

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
	mock.ExpectPutItem().ToTable(test_table)

	service := NewProviderFeedService(test_table, mock_dynamodb, mockHttpClient)

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
	mock.ExpectPutItem().ToTable(test_table)

	service := NewProviderFeedService(test_table, mock_dynamodb, mockHttpClient)

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

func TestCheckForNewVersions(t *testing.T) {
	mock.ExpectGetItem().ToTable(test_table).WithKeys(expectedKey).WillReturns(expectedReturn)
	mock.ExpectPutItem().ToTable(test_table)

	service := NewProviderFeedService(test_table, mock_dynamodb, mockHttpClient)
	provider_versions, err := service.CheckForNewVersions(ProviderAtomFeed{
		ProviderName: test_item_key_value,
		AtomURL:      test_atom_url,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected_len := 4
	actual_len := len(provider_versions)
	if actual_len != expected_len {
		t.Fatalf("expected %d, got %d", expected_len, actual_len)
	}

	expected_latest_version := "v3.7.0"
	actual_latest_version := provider_versions[0].Version
	if actual_latest_version != expected_latest_version {
		t.Fatalf("expected %s, got %s", expected_latest_version, actual_latest_version)
	}

}

func TestCheckForNewVersionsNoEtag(t *testing.T) {
	// ETag: W/"3fd5cf340c5e30202eca209855b7544a"
	mock.ExpectGetItem().ToTable(test_table).WithKeys(expectedKey).WillReturns(dynamodb.GetItemOutput{
		Item: map[string]*dynamodb.AttributeValue{
			"Provider": {S: aws.String(test_item_key_value)},
			"Version":  {S: aws.String(test_item_record.version)},
		},
	})
	mock.ExpectPutItem().ToTable(test_table)

	service := NewProviderFeedService(test_table, mock_dynamodb, mockHttpClient)
	provider_versions, err := service.CheckForNewVersions(ProviderAtomFeed{
		ProviderName: test_item_key_value,
		AtomURL:      test_atom_url,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected_len := 4
	actual_len := len(provider_versions)
	if actual_len != expected_len {
		t.Fatalf("expected %d, got %d", expected_len, actual_len)
	}
}

func TestCheckForNewVersionsSameEtag(t *testing.T) {
	// ETag: W/"3fd5cf340c5e30202eca209855b7544a"
	mock.ExpectGetItem().ToTable(test_table).WithKeys(expectedKey).WillReturns(dynamodb.GetItemOutput{
		Item: map[string]*dynamodb.AttributeValue{
			"Provider": {S: aws.String(test_item_key_value)},
			"ETag":     {S: aws.String(`W/"3fd5cf340c5e30202eca209855b7544a"`)},
			"Version":  {S: aws.String(test_item_record.version)},
		},
	})
	mock.ExpectPutItem().ToTable(test_table)

	service := NewProviderFeedService(test_table, mock_dynamodb, mockHttpClient)
	provider_versions, err := service.CheckForNewVersions(ProviderAtomFeed{
		ProviderName: test_item_key_value,
		AtomURL:      test_atom_url,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected_len := 0
	actual_len := len(provider_versions)
	if actual_len != expected_len {
		t.Fatalf("expected %d, got %d", expected_len, actual_len)
	}
}

func TestCheckForNewVersionsSinceTwoVersionsAgo(t *testing.T) {
	// since v3.5.0
	mock.ExpectGetItem().ToTable(test_table).WithKeys(expectedKey).WillReturns(dynamodb.GetItemOutput{
		Item: map[string]*dynamodb.AttributeValue{
			"Provider": {S: aws.String(test_item_key_value)},
			"ETag":     {S: aws.String(test_item_record.etag)},
			"Version":  {S: aws.String(`v3.5.0`)},
		},
	})
	mock.ExpectPutItem().ToTable(test_table)

	service := NewProviderFeedService(test_table, mock_dynamodb, mockHttpClient)
	provider_versions, err := service.CheckForNewVersions(ProviderAtomFeed{
		ProviderName: test_item_key_value,
		AtomURL:      test_atom_url,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected_len := 2
	actual_len := len(provider_versions)
	if actual_len != expected_len {
		t.Fatalf("expected %d, got %d", expected_len, actual_len)
	}

	expected_latest_version := "v3.7.0"
	actual_latest_version := provider_versions[0].Version
	if actual_latest_version != expected_latest_version {
		t.Fatalf("expected %s, got %s", expected_latest_version, actual_latest_version)
	}

}

func TestCheckForNewVersionsNoVersions(t *testing.T) {
	// since v3.5.0
	mock.ExpectGetItem().ToTable(test_table).WithKeys(expectedKey).WillReturns(dynamodb.GetItemOutput{
		Item: map[string]*dynamodb.AttributeValue{
			"Provider": {S: aws.String(test_item_key_value)},
		},
	})
	mock.ExpectPutItem().ToTable(test_table)

	service := NewProviderFeedService(test_table, mock_dynamodb, mockHttpClient)
	provider_versions, err := service.CheckForNewVersions(ProviderAtomFeed{
		ProviderName: test_item_key_value,
		AtomURL:      test_atom_url,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected_len := 4
	actual_len := len(provider_versions)
	if actual_len != expected_len {
		t.Fatalf("expected %d, got %d", expected_len, actual_len)
	}

	expected_latest_version := "v3.7.0"
	actual_latest_version := provider_versions[0].Version
	if actual_latest_version != expected_latest_version {
		t.Fatalf("expected %s, got %s", expected_latest_version, actual_latest_version)
	}

}
