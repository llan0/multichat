package common

import "net/http"

// for testable HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// wraps http.Client to implement HTTPClient
type HTTPClientAdapter struct {
	Client *http.Client
}

func (a *HTTPClientAdapter) Do(req *http.Request) (*http.Response, error) {
	return a.Client.Do(req)
}

// return a default HTTP client
func DefaultHTTPClient() HTTPClient {
	return &HTTPClientAdapter{Client: http.DefaultClient}
}
