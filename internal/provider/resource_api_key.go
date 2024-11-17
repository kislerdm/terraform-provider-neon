package provider

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func resourceAPIKey() *schema.Resource {
	return &schema.Resource{
		Description: `A key to access the Neon API.

!> Note that the resource does not support import.
`,
		SchemaVersion: 1,
		CreateContext: resourceAPIKeyCreateRetry,
		ReadContext:   resourceAPIKeyReadRetry,
		DeleteContext: resourceAPIKeyDeleteRetry,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the API Key.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The API key ID.",
			},
			"key": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The generated 64-bit token required to access the Neon API.",
			},
		},
	}
}

type apiKey struct {
	id      string
	keyName string
	key     *string
}

func updateStateAPIKey(d *schema.ResourceData, v apiKey) error {
	d.SetId(v.id)
	if err := d.Set("name", v.keyName); err != nil {
		return err
	}
	if v.key != nil {
		if err := d.Set("key", *v.key); err != nil {
			return err
		}
	}
	return nil
}

func resourceAPIKeyCreateRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceAPIKeyCreate, ctx, d, meta)
}

func resourceAPIKeyCreate(_ context.Context, d *schema.ResourceData, meta interface{}) error {
	keyName := d.Get("name").(string)
	resp, err := meta.(*neon.Client).CreateApiKey(
		neon.ApiKeyCreateRequest{
			KeyName: keyName,
		},
	)
	if err != nil {
		return err
	}

	return updateStateAPIKey(d, apiKey{
		id:      strconv.FormatInt(resp.ID, 10),
		key:     &resp.Key,
		keyName: keyName,
	})
}

func resourceAPIKeyReadRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceAPIKeyRead, ctx, d, meta)
}

func resourceAPIKeyRead(_ context.Context, d *schema.ResourceData, meta interface{}) error {
	resp, err := meta.(*neon.Client).ListApiKeys()
	if err != nil {
		return err
	}

	keyName := d.Get("name").(string)

	var id int64

	found := slices.ContainsFunc(resp, func(key neon.ApiKeysListResponseItem) bool {
		if keyName == key.Name {
			id = key.ID
		}
		return keyName == key.Name
	})

	if !found {
		return fmt.Errorf("couldn't find API Key %s", keyName)
	}

	return updateStateAPIKey(d, apiKey{
		id:      strconv.FormatInt(id, 10),
		keyName: keyName,
	})
}

func resourceAPIKeyDeleteRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceAPIKeyDelete, ctx, d, meta)
}

func resourceAPIKeyDelete(_ context.Context, d *schema.ResourceData, meta interface{}) error {
	id, err := strconv.ParseInt(d.Get("id").(string), 10, 64)
	if err != nil {
		return err
	}

	if _, err := meta.(*neon.Client).RevokeApiKey(id); err != nil {
		return err
	}

	emptyStr := ""
	return updateStateAPIKey(d, apiKey{key: &emptyStr})
}
