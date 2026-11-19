// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package listresources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/thrashr888/terraform-provider-registry/internal/client"
)

// getJSON fetches reqURL with the provider's client and decodes the JSON body into out.
func getJSON(ctx context.Context, c *client.RegistryClient, reqURL string, out any) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}
	if c.Token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.Token)
	}

	httpResp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", httpResp.StatusCode)
	}
	return json.NewDecoder(httpResp.Body).Decode(out)
}
