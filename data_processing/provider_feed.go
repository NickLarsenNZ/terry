package data_processing

import (
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/mmcdole/gofeed"
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
	http       *http.Client
	table      string
	cache      map[string]DynamoDBMapAttributeValue
	cache_hit  int64
	cache_miss int64
}

func NewProviderFeedService(table_name string, dynamodb dynamodbiface.DynamoDBAPI, http *http.Client) *ProviderFeedService {
	return &ProviderFeedService{
		dynamodb: dynamodb,
		http:     http,
		table:    table_name,
		cache:    map[string]DynamoDBMapAttributeValue{},
	}
}

type ProviderVersion struct {
	Provider string
	Version  string
}

func (s *ProviderFeedService) CheckForNewVersions(input ProviderAtomFeed) ([]ProviderVersion, error) {
	var newVersions []ProviderVersion

	lastVersion, err := s.LastVersion(input.ProviderName)
	if err != nil && !errors.Is(err, ErrItemNotFound) {
		return nil, err
	} else {
		// Otherwise, get the last seen ETag, and check (HEAD) if there are any updates to the feed
		lastETag, err := s.LastETag(input.ProviderName)
		if err != nil && !errors.Is(err, ErrItemNotFound) {
			return nil, err
		} else {
			// Otherwise check the new ETag
			res, err := s.http.Head(input.AtomURL)
			if err != nil {
				return nil, err
			}

			newETag := res.Header.Get("ETag")
			if newETag == lastETag {
				// No new versions, return empty result
				return []ProviderVersion{}, nil
			}
		}
	}

	// If no Version, no ETag, or different to what we had, fetch the updates
	res, err := s.http.Get(input.AtomURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	newETag := res.Header.Get("ETag")
	fp := gofeed.NewParser()
	feed, err := fp.Parse(res.Body)
	versions := s.getVersionsSince(feed, lastVersion)
	newVersion := versions[0]

	// Update the last ETag and Version in the DB
	_ = newETag
	_ = newVersion

	// return
	for _, version := range versions {
		newVersions = append(newVersions, ProviderVersion{
			Provider: input.ProviderName,
			Version:  version,
		})
	}

	return newVersions, nil
}

func (s *ProviderFeedService) getVersionsSince(feed *gofeed.Feed, sinceVersion string) []string {
	var versions []string
	for _, item := range feed.Items {
		bits := strings.Split(item.GUID, "/")
		tag := bits[len(bits)-1]
		if tag[0] != 'v' {
			continue
		}

		if sinceVersion != "" && sinceVersion == tag {
			break
		}

		versions = append(versions, tag)
	}
	return versions
}

func (s *ProviderFeedService) LastVersion(provider_name string) (string, error) {
	item, err := s.getItem(provider_name)
	if err != nil || item["Version"] == nil {
		return "", err
	}

	return *item["Version"].S, nil
}

func (s *ProviderFeedService) LastETag(provider_name string) (string, error) {
	item, err := s.getItem(provider_name)
	if err != nil || item["ETag"] == nil {
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
