// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"fmt"
	"net/http"
	"time"
)

// DefaultTimeout bounds every registry request so a stalled connection
// cannot hang a plan or apply.
const DefaultTimeout = 30 * time.Second

// RegistryClient is the API client for Terraform Registry.
type RegistryClient struct {
	Host       string
	Token      string
	HTTPClient *http.Client
}

// New returns a RegistryClient whose HTTP client has a timeout and sends a
// User-Agent naming this provider and its version on every request.
func New(host, token, version string) *RegistryClient {
	return &RegistryClient{
		Host:  host,
		Token: token,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
			Transport: &userAgentTransport{
				userAgent: UserAgent(version),
				base:      http.DefaultTransport,
			},
		},
	}
}

// UserAgent returns the User-Agent header value for the given provider version.
func UserAgent(version string) string {
	if version == "" {
		version = "dev"
	}
	return fmt.Sprintf("terraform-provider-registry/%s (+https://github.com/thrashr888/terraform-provider-registry)", version)
}

// BaseURL returns the base URL for the registry API.
func (c *RegistryClient) BaseURL() string {
	return "https://" + c.Host + "/v1"
}

type userAgentTransport struct {
	userAgent string
	base      http.RoundTripper
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get("User-Agent") == "" {
		req = req.Clone(req.Context())
		req.Header.Set("User-Agent", t.userAgent)
	}
	return t.base.RoundTrip(req)
}
