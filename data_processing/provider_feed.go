package data_processing

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/pkg/errors"
)

var (
	ErrETagNotFound = errors.New("ETag Not Found")
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

func (s *ProviderFeedService) LastETag(provider_name string) (string, error) {
	params := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"Provider": {
				S: aws.String(provider_name),
			},
		},
		TableName: aws.String(s.table),
	}

	resp, err := s.dynamodb.GetItem(params)
	if err != nil {
		return "", ErrETagNotFound
	}

	lastETag := resp.Item["ETag"].S

	return *lastETag, nil
}
