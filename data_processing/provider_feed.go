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

type ProviderFeedService struct {
	dynamodb dynamodbiface.DynamoDBAPI
	table    string
}

func NewProviderFeedService(table_name string, dynamodb dynamodbiface.DynamoDBAPI) *ProviderFeedService {
	return &ProviderFeedService{
		dynamodb: dynamodb,
		table:    table_name,
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

	return resp.Item, nil
}
