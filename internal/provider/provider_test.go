// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"
)

func TestProvider_Metadata(t *testing.T) {
	// Basic test to ensure provider loads
	provider := New("test")()
	if provider == nil {
		t.Fatal("expected provider, got nil")
	}
}
