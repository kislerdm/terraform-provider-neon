package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func resourceAPIKey() *schema.Resource {
	return &schema.Resource{
		Description:   "A key to access the Neon API.",
		SchemaVersion: 7,
		CreateContext: resourceAPIKeyCreateRetry,
		ReadContext:   resourceAPIKeyReadRetry,
		DeleteContext: resourceAPIKeyDeleteRetry,
		Schema: map[string]*schema.Schema{
			"key_name": {
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
	key     string
	keyName string
}

func updateStateAPIKey(d *schema.ResourceData, v apiKey) error {
	if err := d.Set("key_name", v.keyName); err != nil {
		return err
	}
	if err := d.Set("id", v.id); err != nil {
		return err
	}
	if v.key != "" {
		if err := d.Set("key", v.key); err != nil {
			return err
		}
	}
	return nil
}

func resourceAPIKeyCreateRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceAPIKeyCreate, ctx, d, meta)
}

func resourceAPIKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	keyName := d.Get("key_name").(string)
	resp, err := meta.(*neon.Client).CreateApiKey(
		neon.ApiKeyCreateRequest{
			KeyName: keyName,
		},
	)
	if err != nil {
		return err
	}

	key := resp.Key
	id := strconv.Itoa(int(resp.ID))
	d.SetId(id)

	return updateStateAPIKey(d, apiKey{
		id:      id,
		key:     key,
		keyName: keyName,
	})
}

func resourceAPIKeyReadRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceAPIKeyRead, ctx, d, meta)
}

func resourceAPIKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	keyName := d.Get("key_name").(string)
	resp, err := meta.(*neon.Client).ListApiKeys()
	if err != nil {
		return err
	}

	found := false
	var id int64
	for _, key := range resp {
		if keyName == key.Name {
			found = true
			id = key.ID
		}
	}

	if !found {
		return fmt.Errorf("couldn't find API Key %s", keyName)
	}

	return updateStateAPIKey(d, apiKey{
		id:      strconv.Itoa(int(id)),
		keyName: keyName,
	})
}

func resourceAPIKeyDeleteRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceAPIKeyDelete, ctx, d, meta)
}

func resourceAPIKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	id, err := strconv.Atoi(d.Get("id").(string))
	if err != nil {
		return err
	}

	if _, err := meta.(*neon.Client).RevokeApiKey(int64(id)); err != nil {
		return err
	}

	d.SetId("")
	if err := d.Set("key_name", ""); err != nil {
		return err
	}
	if err := d.Set("id", ""); err != nil {
		return err
	}

	return updateStateAPIKey(d, apiKey{})
}
