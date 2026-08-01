package provider

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	neon "github.com/kislerdm/neon-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestRecreateAPIKeyIfNotFound(t *testing.T) {
	// see: https://github.com/kislerdm/terraform-provider-neon/issues/209

	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1")
	}

	client, err := neon.NewClient(neon.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		t.Fatal(err)
	}

	var preConfig = func(name string) {
		ref, err := client.ListApiKeys()
		if err != nil {
			panic(err)
		}

		for _, key := range ref {
			if key.Name == name {
				_, err = client.RevokeApiKey(key.ID)
				if err != nil {
					panic(err)
				}
			}
		}

	}

	t.Cleanup(func() {
		ref, err := client.ListApiKeys()
		if err != nil {
			panic(err)
		}

		for _, key := range ref {
			if strings.HasPrefix(key.Name, "test") {
				_, err = client.RevokeApiKey(key.ID)
				if err != nil {
					panic(err)
				}
			}
		}
	})

	t.Run("shall indicate non empty plan if the API key was deleted outside of terraform", func(t *testing.T) {
		keyName := "test" + uuid.NewString()
		resource.Test(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return newAccTest(), nil
					},
				},
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`resource "neon_api_key" "this" {name = "%s"}`, keyName),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"neon_api_key.this",
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
		keyName := "test" + uuid.NewString()
		config := fmt.Sprintf(`resource "neon_api_key" "this" {name = "%s"}`, keyName)
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
								"neon_api_key.this",
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
							_, ok := s.RootModule().Resources["neon_api_key.this"]
							assert.False(t, ok, "resource neon_api_key.this should be destroyed")
							return nil
						},
					},
				},
			})
	})
}
