package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	DefaultMaximumBackoffElapsedTime = 10 * time.Second
)

type Client struct {
	httpClient                *http.Client
	URL                       string
	maximumBackoffElapsedTime time.Duration
}

func NewClient(client *http.Client, url string, maximumBackoffElapsedTime time.Duration) *Client {
	if maximumBackoffElapsedTime <= 0 {
		maximumBackoffElapsedTime = DefaultMaximumBackoffElapsedTime
	}

	underlyingTransport := http.DefaultTransport
	if client.Transport != nil {
		underlyingTransport = client.Transport
	}

	// wrap instrumentation around the existing http.Client transport
	client.Transport = promhttp.InstrumentRoundTripperInFlight(RequestsToDiscordInFlight,
		promhttp.InstrumentRoundTripperCounter(RequestsToDiscordTotal,
			promhttp.InstrumentRoundTripperDuration(RequestsToDiscordDuration,
				underlyingTransport,
			),
		),
	)

	return &Client{
		httpClient:                client,
		URL:                       url,
		maximumBackoffElapsedTime: maximumBackoffElapsedTime,
	}
}

func (dc *Client) PublishMessage(message Out) (*http.Response, error) {
	DOD, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("Error encountered when marshalling object to json. We will not continue posting to Discord. Discord Out object: '%v+'. Error: %w", message, err)
	}

	var response *http.Response

	operation := func() error {
		res, err := dc.httpClient.Post(dc.URL, "application/json", bytes.NewReader(DOD))
		if err == nil {
			response = res
		}

		return err
	}

	exponential := backoff.NewExponentialBackOff()
	exponential.MaxElapsedTime = dc.maximumBackoffElapsedTime
	err = backoff.Retry(operation, exponential)
	if err != nil {
		return nil, fmt.Errorf("Error encountered sending POST to '%s'. Error: %w", dc.URL, err)
	}

	return response, nil
}
