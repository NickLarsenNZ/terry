package data_processing

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type ProviderFeedsService struct {
	http      *http.Client
	feeds_url string
}

type ProviderAtomFeed struct {
	ProviderName string `json:"provider_name"`
	AtomURL      string `json:"atom_url"`
}

func NewProviderFeedsService(feeds_url string, http *http.Client) *ProviderFeedsService {
	// Should ensure the URL is valid
	u, err := url.ParseRequestURI(feeds_url)
	if err != nil {
		fmt.Printf("Feeds URL is not valid: %s\n", feeds_url)
		panic(err)
	}

	return &ProviderFeedsService{
		http:      http,
		feeds_url: u.String(),
	}
}

func (s *ProviderFeedsService) GetProviderFeeds() ([]ProviderAtomFeed, error) {
	res, err := s.http.Get(s.feeds_url)
	if err != nil {
		return nil, err
	}

	var feeds []ProviderAtomFeed

	if res.StatusCode != 200 {
		defer res.Body.Close()
		b, _ := ioutil.ReadAll(res.Body)
		return nil, errors.New(string(b))
	}

	err = json.NewDecoder(res.Body).Decode(&feeds)
	return feeds, err
}
