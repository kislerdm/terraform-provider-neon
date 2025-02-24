package provider

import (
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	neon "github.com/kislerdm/neon-sdk-go"
)

func TestAccAPIKeys(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1")
	}

	client, err := neon.NewClient(neon.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		t.Fatal(err)
	}

	wantKeyName := projectNamePrefix + "apikey"

	t.Cleanup(func() {
		keys, _ := client.ListApiKeys()
		for _, key := range keys {
			if wantKeyName == key.Name {
				_, _ = client.RevokeApiKey(key.ID)
			}
		}
	})

	resourceDefinition := fmt.Sprintf(`resource "neon_api_key" "this" { name = "%s" }`, wantKeyName)
	const resourceName = "neon_api_key.this"

	resource.Test(
		t, resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"neon": func() (*schema.Provider, error) {
					return newAccTest(), nil
				},
			},
			Steps: []resource.TestStep{
				{
					Config: resourceDefinition,
					Check: resource.ComposeTestCheckFunc(
						// verify the recorded key name
						resource.TestCheckResourceAttr(resourceName, "name", wantKeyName),
						// verify that the key with the given name was created
						func(_ *terraform.State) error {
							keys, e := client.ListApiKeys()
							if e == nil {
								if !slices.ContainsFunc(keys, func(key neon.ApiKeysListResponseItem) bool {
									return wantKeyName == key.Name
								}) {
									e = fmt.Errorf("key %s not found", wantKeyName)
								}
							}
							return e
						},
						// verify that the valid key was recorded
						resource.TestCheckResourceAttrWith(resourceName, "key", func(value string) error {
							_, err := neon.NewClient(neon.Config{Key: value})
							return err
						}),
					),
				},
				{
					Config:  resourceDefinition,
					Destroy: true,
					Check: resource.ComposeTestCheckFunc(
						// verify that the key with the given name was indeed revoked
						func(_ *terraform.State) error {
							keys, e := client.ListApiKeys()
							if e == nil {
								if slices.ContainsFunc(keys, func(key neon.ApiKeysListResponseItem) bool {
									return wantKeyName == key.Name
								}) {
									e = fmt.Errorf("key %s is expected to be not found", wantKeyName)
								}
							}

							return e
						},
					),
				},
			},
		})
}
