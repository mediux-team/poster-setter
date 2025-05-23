package utils

import (
	"aura/internal/logging"
	"fmt"
	"net/url"
	"path"
)

func MakeMediaServerAPIURL(endpoint, baseURL string) (*url.URL, logging.ErrorLog) {
	// Check if the base URL is empty
	if baseURL == "" {
		return nil, logging.ErrorLog{Err: fmt.Errorf("base URL is empty"),
			Log: logging.Log{Message: "Base URL is empty"}}
	}

	// Parse the base URL
	baseURLParsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, logging.ErrorLog{Err: err,
			Log: logging.Log{Message: "Error parsing base URL"}}
	}

	// Construct the full URL by appending the endpoint
	baseURLParsed.Path = path.Join(baseURLParsed.Path, endpoint)

	return baseURLParsed, logging.ErrorLog{}
}
