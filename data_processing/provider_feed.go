package data_processing

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/pkg/errors"
)

var (
	ErrItemNotFound = errors.New("Item Not Found")
)

// Because I can't seem to initialise a `map[string]map[string]*dynamodb.AttributeValue`
// Either by `make(map[string]map[string]*dynamodb.AttributeValue)` or `map[string]map[string]*dynamodb.AttributeValue{}`
type DynamoDBMapAttributeValue map[string]*dynamodb.AttributeValue

type ProviderFeedService struct {
	dynamodb   dynamodbiface.DynamoDBAPI
	table      string
	cache      map[string]DynamoDBMapAttributeValue
	cache_hit  int64
	cache_miss int64
}

func NewProviderFeedService(table_name string, dynamodb dynamodbiface.DynamoDBAPI) *ProviderFeedService {
	return &ProviderFeedService{
		dynamodb: dynamodb,
		table:    table_name,
		cache:    map[string]DynamoDBMapAttributeValue{},
	}
}

func (s *ProviderFeedService) LastVersion(provider_name string) (string, error) {
	item, err := s.getItem(provider_name)
	if err != nil {
		return "", err
	}

	return *item["Version"].S, nil
}

func (s *ProviderFeedService) LastETag(provider_name string) (string, error) {
	item, err := s.getItem(provider_name)
	if err != nil {
		return "", err
	}

	return *item["ETag"].S, nil
}

func (s *ProviderFeedService) getItem(key string) (map[string]*dynamodb.AttributeValue, error) {
	// If we have already retrieved it, then don't bother making the call again
	if s.cache[key] != nil {
		s.cache_hit += 1
		return s.cache[key], nil
	}

	params := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"Provider": {
				S: aws.String(key),
			},
		},
		TableName: aws.String(s.table),
	}

	resp, err := s.dynamodb.GetItem(params)
	if err != nil {
		return nil, errors.Wrap(ErrItemNotFound, err.Error())
	}

	// Store in the cache for subsequent calls
	s.cache_miss += 1
	s.cache[key] = resp.Item

	return resp.Item, nil
}
