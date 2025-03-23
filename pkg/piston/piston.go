package piston

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/alvii147/nymphadora-api/pkg/api"
	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/httputils"
)

// mu is a mutex intended to avoid concurrent requests to Piston.
// Requests need to be made at least 200 ms apart to avoid hiting the rate limit.
var mu sync.Mutex

const (
	// PistonRateLimit is the rate limit duration for Piston.
	PistonRateLimit = 200 * time.Millisecond
	// PistonExecuteURL is the endpoint for code execution on Piston.
	PistonExecuteURL = "https://emkc.org/api/v2/piston/execute"
)

// Client represents a client that sends requests to Piston.
//
//go:generate mockgen -package=pistonmocks -source=$GOFILE -destination=./mocks/piston.go
type Client interface {
	Execute(request *api.PistonExecuteRequest) (*api.PistonExecuteResponse, error)
}

// client implements Client.
type client struct {
	key        *string
	httpClient httputils.HTTPClient
}

// NewClient returns a new client.
func NewClient(key *string, httpClient httputils.HTTPClient) *client {
	return &client{
		key:        key,
		httpClient: httpClient,
	}
}

// Execute sends a remote code execution request to Piston.
func (c *client) Execute(data *api.PistonExecuteRequest) (*api.PistonExecuteResponse, error) {
	mu.Lock()
	defer mu.Unlock()

	body, err := json.Marshal(data)
	if err != nil {
		return nil, errutils.FormatError(err, "json.Marshal failed")
	}

	req, err := http.NewRequest(http.MethodPost, PistonExecuteURL, bytes.NewReader(body))
	if err != nil {
		return nil, errutils.FormatError(err, "http.NewRequest failed")
	}

	if c.key != nil {
		req.Header.Set("Authorization", *c.key)
	}

	// sleep to avoid hitting rate limit
	time.Sleep(PistonRateLimit)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errutils.FormatError(err, "c.httpClient.Do failed")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)

		return nil, errutils.FormatErrorf(
			err,
			"c.httpClient.Do returned status code %d %v",
			resp.StatusCode,
			string(bodyBytes),
		)
	}

	results := &api.PistonExecuteResponse{}
	err = json.NewDecoder(resp.Body).Decode(results)
	if err != nil {
		return nil, errutils.FormatError(err, "json.Decoder.Decode failed")
	}

	return results, nil
}
