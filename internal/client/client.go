// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"net/http"
)

// RegistryClient is the API client for Terraform Registry.
type RegistryClient struct {
	Host       string
	Token      string
	HTTPClient *http.Client
}

// BaseURL returns the base URL for the registry API.
func (c *RegistryClient) BaseURL() string {
	return "https://" + c.Host + "/v1"
}
