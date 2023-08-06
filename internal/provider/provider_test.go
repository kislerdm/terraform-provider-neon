package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var providerFactories = map[string]func() (*schema.Provider, error){
	"neon": func() (*schema.Provider, error) {
		return New("0.3.0"), nil
	},
}

func TestProvider(t *testing.T) {
	if err := New("dev").InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("NEON_API_KEY") != "" {
		t.Fatalf("NEON_API_KEY must be set")
	}
}
