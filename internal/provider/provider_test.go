// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"registry": providerserver.NewProtocol6WithError(New("test")()),
}

func TestProvider_Metadata(t *testing.T) {
	// Basic test to ensure provider loads
	provider := New("test")()
	if provider == nil {
		t.Fatal("expected provider, got nil")
	}
}
