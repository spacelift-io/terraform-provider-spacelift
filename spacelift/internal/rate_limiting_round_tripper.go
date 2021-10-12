package internal

import (
	"net/http"

	"github.com/pkg/errors"
	"golang.org/x/time/rate"
)

// rateLimitingRoundTripper is an HTTP round tripper that uses a Limiter to rate limit requests.
type rateLimitingRoundTripper struct {
	client  *http.Client
	limiter *rate.Limiter
}

// newRateLimitingRoundTripper create a new round tripper.
func newRateLimitingRoundTripper(client *http.Client, limiter *rate.Limiter) *rateLimitingRoundTripper {
	return &rateLimitingRoundTripper{
		client:  client,
		limiter: limiter,
	}
}

// RoundTrip executes the specified request.
func (r *rateLimitingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := r.limiter.Wait(req.Context()); err != nil {
		return nil, errors.Wrap(err, "could not get request token from limiter")
	}

	return r.client.Do(req)
}
