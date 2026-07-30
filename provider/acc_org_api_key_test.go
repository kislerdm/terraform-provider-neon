package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	neon "github.com/kislerdm/neon-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestRecreateOrgAPIKeyIfNotFound(t *testing.T) {
	// see: https://github.com/kislerdm/terraform-provider-neon/issues/209

	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1")
	}

	orgID := os.Getenv("ORG_ID")
	if orgID == "" {
		t.Skip("ORG_ID must be set")
	}

	client, err := neon.NewClient(neon.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		t.Fatal(err)
	}

	var preConfig = func(name string) {
		ref, err := client.ListOrgApiKeys(orgID)
		if err != nil {
			panic(err)
		}

		for _, key := range ref {
			if key.Name == name {
				_, err = client.RevokeOrgApiKey(orgID, key.ID)
				if err != nil {
					panic(err)
				}
			}
		}

	}

	t.Run("shall indicate non empty plan if the API key was deleted outside of terraform", func(t *testing.T) {
		keyName := "test"
		resource.Test(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return newAccTest(), nil
					},
				},
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`resource "neon_org_api_key" "this" {
	name   = "%s"
	org_id = "%s"
}`, keyName, orgID),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"neon_org_api_key.this",
								"name", keyName,
							),
						),
					},
					{
						PreConfig: func() {
							preConfig(keyName)
						},
						RefreshState:       true,
						ExpectNonEmptyPlan: true,
					},
				},
			})
	})

	t.Run("shall destroy even if the API key was deleted outside of terraform,", func(t *testing.T) {
		keyName := "test"
		config := fmt.Sprintf(`resource "neon_org_api_key" "this" {
	name   = "%s"
	org_id = "%s"
}`, keyName, orgID)
		resource.Test(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return newAccTest(), nil
					},
				},
				Steps: []resource.TestStep{
					{
						Config: config,
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"neon_org_api_key.this",
								"name", keyName,
							),
						),
					},
					{
						PreConfig: func() {
							preConfig(keyName)
						},
						Config:  config,
						Destroy: true,
						Check: func(s *terraform.State) error {
							_, ok := s.RootModule().Resources["neon_org_api_key.this"]
							assert.False(t, ok, "resource neon_org_api_key.this should be destroyed")
							return nil
						},
					},
				},
			})
	})
}
