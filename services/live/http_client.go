package main

import (
	"net/http"
	"time"
)

// internalHTTPClient is shared by service-to-service calls. A bounded timeout
// prevents a stalled dependency from pinning live requests indefinitely while
// the transport keeps connections reusable as the service scales horizontally.
var internalHTTPClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}
