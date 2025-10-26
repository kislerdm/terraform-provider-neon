package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func resourceOrgAPIKey() *schema.Resource {
	return &schema.Resource{
		Description: `An org-specific key to access the Neon API.

~>**WARNING** The resource does not support import.
`,
		SchemaVersion: 1,
		CreateContext: resourceOrgAPIKeyCreateRetry,
		ReadContext:   resourceOrgAPIKeyReadRetry,
		DeleteContext: resourceOrgAPIKeyDeleteRetry,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the API Key.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The organisation ID.",
			},
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The project ID to which this key will grant the access to.",
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

func resourceOrgAPIKeyCreateRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceOrgAPIKeyCreate, ctx, d, meta)
}

func resourceOrgAPIKeyCreate(_ context.Context, d *schema.ResourceData, meta interface{}) error {
	req := neon.OrgApiKeyCreateRequest{
		ApiKeyCreateRequest: neon.ApiKeyCreateRequest{
			KeyName: d.Get("name").(string),
		},
	}
	if v, ok := d.GetOk("project_id"); ok {
		s := v.(string)
		req.ProjectID = &s
	}
	resp, err := meta.(*neon.Client).CreateOrgApiKey(
		d.Get("org_id").(string),
		req,
	)
	if err != nil {
		return err
	}
	d.SetId(strconv.FormatInt(resp.ID, 10))
	return d.Set("key", resp.Key)
}

func resourceOrgAPIKeyReadRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceOrgAPIKeyRead, ctx, d, meta)
}

func resourceOrgAPIKeyRead(_ context.Context, d *schema.ResourceData, meta interface{}) error {
	resp, err := meta.(*neon.Client).ListOrgApiKeys(d.Get("org_id").(string))

	if err == nil {
		keyName := d.Get("name").(string)

		var found bool
		for _, v := range resp {
			if keyName == v.Name {
				d.SetId(strconv.FormatInt(v.ID, 10))
				found = true
				break
			}
		}

		if !found {
			err = fmt.Errorf("couldn't find API Key %s", keyName)
		}
	}

	return err
}

func resourceOrgAPIKeyDeleteRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceOrgAPIKeyDelete, ctx, d, meta)
}

func resourceOrgAPIKeyDelete(_ context.Context, d *schema.ResourceData, meta interface{}) error {
	id, err := strconv.ParseInt(d.Get("id").(string), 10, 64)
	if err != nil {
		return err
	}

	if _, err := meta.(*neon.Client).RevokeOrgApiKey(d.Get("org_id").(string), id); err != nil {
		return err
	}

	if err = d.Set("key", ""); err != nil {
		return err
	}
	if err = d.Set("name", ""); err != nil {
		return err
	}
	if err = d.Set("project_id", ""); err != nil {
		return err
	}
	if err = d.Set("org_id", ""); err != nil {
		return err
	}
	d.SetId("")
	return nil
}
