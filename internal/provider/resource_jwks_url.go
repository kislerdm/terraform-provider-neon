package provider

import (
	"context"
	"errors"
	"slices"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func resourceJwksUrl() *schema.Resource {
	return &schema.Resource{
		Description: `Project JWKS URL. See details: https://neon.tech/docs/guides/neon-rls-authorize
`,
		SchemaVersion: 1,
		Importer: &schema.ResourceImporter{
			StateContext: resourceJwksUrlImport,
		},
		CreateContext: resourceJwksUrlCreateRetry,
		ReadContext:   resourceJwksUrlReadRetry,
		DeleteContext: resourceJwksUrlDeleteRetry,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Project ID.",
			},
			"jwks_url": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The URL that lists the JWKS.",
			},
			"provider_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the authentication provider.",
			},
			"role_names": {
				Type:        schema.TypeList,
				MinItems:    1,
				MaxItems:    10,
				Required:    true,
				ForceNew:    true,
				Description: "The roles the JWKS should be mapped to.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"branch_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Branch ID.",
			},
			"jwt_audience": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the required JWT Audience to be used.",
			},
		},
	}
}

func updateStateJwksUrl(d *schema.ResourceData, v neon.JWKS, role_names []string) error {
	if err := d.Set("project_id", v.ProjectID); err != nil {
		return err
	}
	if err := d.Set("jwks_url", v.JwksURL); err != nil {
		return err
	}
	if err := d.Set("provider_name", v.ProviderName); err != nil {
		return err
	}
	// todo - we don't get this back from the API, is using the provided values ok?
	if err := d.Set("role_names", role_names); err != nil {
		return err
	}
	if err := d.Set("branch_id", v.BranchID); err != nil {
		return err
	}
	if err := d.Set("jwt_audience", v.JwtAudience); err != nil {
		return err
	}

	return nil
}

func resourceJwksUrlCreateRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceJwksUrlCreate, ctx, d, meta)
}

func resourceJwksUrlCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "created JWKS URL")

	cfg := neon.AddProjectJWKSRequest{
		JwksURL:      d.Get("jwks_url").(string),
		ProviderName: d.Get("provider_name").(string),
		RoleNames:    d.Get("role_names").([]string),
		BranchID:     pointer(d.Get("branch_id").(string)),
		JwtAudience:  pointer(d.Get("jwt_audience").(string)),
	}

	client := meta.(*neon.Client)
	resp, err := client.AddProjectJWKS(d.Get("project_id").(string), cfg)
	if err != nil {
		return err
	}
	waitUnfinishedOperations(ctx, client, resp.OperationsResponse.Operations)

	d.SetId(resp.Jwks.ID)

	return updateStateJwksUrl(d, resp.JWKSResponse.Jwks, cfg.RoleNames)
}

func resourceJwksUrlReadRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceJwksUrlRead, ctx, d, meta)
}

func resourceJwksUrlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "read JWKS URL")

	projectID, _ := d.Get("project_id").(string)
	id, _ := d.Get("id").(string)

	resp, err := meta.(*neon.Client).GetProjectJWKS(projectID)
	if err != nil {
		return err
	}

	idx := slices.IndexFunc(resp.Jwks, func(j neon.JWKS) bool { return j.ID == id })
	if idx < 0 {
		return errors.New("JWKS URL resource was not found")
	}
	jwks := resp.Jwks[idx]

	return updateStateJwksUrl(d, jwks, []string{})
}

func resourceJwksUrlDeleteRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceJwksUrlDelete, ctx, d, meta)
}

func resourceJwksUrlDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "delete JWKS URL")
	client := meta.(*neon.Client)
	resp, err := client.DeleteProjectJWKS(
		d.Get("project_id").(string),
		d.Get("id").(string),
	)
	if err != nil {
		return err
	}
	//todo - the response object doesn't have an OperationsResponse attribute
	waitUnfinishedOperations(ctx, client, resp.OperationsResponse.Operations)
	d.SetId("")
	return updateStateJwksUrl(d, neon.JWKS{}, []string{})
}

func resourceJwksUrlImport(ctx context.Context, d *schema.ResourceData, meta interface{}) (
	[]*schema.ResourceData, error,
) {
	tflog.Trace(ctx, "import JWKS URL")

	if diags := resourceJwksUrlReadRetry(ctx, d, meta); diags.HasError() {
		return nil, errors.New(diags[0].Summary)
	}
	return []*schema.ResourceData{d}, nil
}
