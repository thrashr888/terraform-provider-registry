// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewSetsTimeoutAndUserAgent(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("User-Agent")
	}))
	defer srv.Close()

	c := New("registry.terraform.io", "", "1.2.3")
	if c.HTTPClient.Timeout != DefaultTimeout {
		t.Fatalf("timeout = %v, want %v", c.HTTPClient.Timeout, DefaultTimeout)
	}
	resp, err := c.HTTPClient.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if want := UserAgent("1.2.3"); got != want {
		t.Fatalf("User-Agent = %q, want %q", got, want)
	}
}

func TestUserAgentDefaultsToDev(t *testing.T) {
	if got := UserAgent(""); got != UserAgent("dev") {
		t.Fatalf("UserAgent(\"\") = %q", got)
	}
}
