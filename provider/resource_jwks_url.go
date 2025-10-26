package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func resourceJwksUrl() *schema.Resource {
	return &schema.Resource{
		Description: `Project JWKS URL. See details: https://neon.tech/docs/guides/neon-rls-authorize

~>**WARNING** The resource does not support import.
`,
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
				ForceNew:    true,
				Optional:    true,
				Description: "Branch ID.",
			},
			"jwt_audience": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "The name of the required JWT Audience to be used.",
			},
		},
	}
}

func updateStateJwksUrl(d *schema.ResourceData, v neon.JWKS, roleNames *[]string) error {
	if err := d.Set("project_id", v.ProjectID); err != nil {
		return err
	}
	if err := d.Set("jwks_url", v.JwksURL); err != nil {
		return err
	}
	if err := d.Set("provider_name", v.ProviderName); err != nil {
		return err
	}
	if roleNames != nil {
		if err := d.Set("role_names", roleNames); err != nil {
			return err
		}
	}
	if v.BranchID != nil {
		if err := d.Set("branch_id", *v.BranchID); err != nil {
			return err
		}
	}
	if v.JwtAudience != nil {
		if err := d.Set("jwt_audience", *v.JwtAudience); err != nil {
			return err
		}
	}
	return nil
}

func resourceJwksUrlCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "create JWKS URL")
	cfg := neon.AddProjectJWKSRequest{
		JwksURL:      d.Get("jwks_url").(string),
		ProviderName: d.Get("provider_name").(string),
	}
	// The length condition is an overhead because tf type system will verify that the slice contains at least 1 element
	// I decided to keep it as additional gateway in case the API interface changes.
	if v, ok := d.GetOk("role_names"); ok && len(v.([]interface{})) > 0 {
		vv := v.([]interface{})
		var roleNames = make([]string, len(vv))
		for i, roleName := range vv {
			roleNames[i] = roleName.(string)
		}
		cfg.RoleNames = &roleNames
	}
	if v, ok := d.GetOk("branch_id"); ok && v.(string) != "" {
		cfg.BranchID = pointer(v.(string))
	}
	if v, ok := d.GetOk("jwt_audience"); ok && v.(string) != "" {
		cfg.JwtAudience = pointer(v.(string))
	}

	tflog.Debug(ctx, "create JWKS URL", map[string]interface{}{"cfg": cfg})

	client := meta.(*neon.Client)
	resp, err := client.AddProjectJWKS(d.Get("project_id").(string), cfg)
	if err == nil {
		waitUnfinishedOperations(ctx, client, resp.OperationsResponse.Operations)
		err = updateStateJwksUrl(d, resp.JWKSResponse.Jwks, cfg.RoleNames)
	}
	if err == nil {
		d.SetId(resp.Jwks.ID)
		tflog.Trace(ctx, "successfully created JWKS URL")
	}
	return err
}

func resourceJwksUrlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	var err error
	tflog.Trace(ctx, "read JWKS URL", map[string]interface{}{"id": d.Id()})
	projectID, ok := d.Get("project_id").(string)
	if !ok {
		err = errors.New("project_id is not a string")
	}

	var resp neon.ProjectJWKSResponse
	if err == nil {
		resp, err = meta.(*neon.Client).GetProjectJWKS(projectID)
	}
	if err == nil {
		var jwks neon.JWKS
		for _, el := range resp.Jwks {
			if d.Id() == el.ID {
				jwks = el
				break
			}
		}
		if jwks.ID == "" {
			err = fmt.Errorf("could not find JWKS %s for project %s", d.Id(), projectID)
		} else {
			err = updateStateJwksUrl(d, jwks, nil)
		}
	}
	tflog.Trace(ctx, "successfully read JWKS URL", map[string]interface{}{"id": d.Id()})
	return err
}

func resourceJwksUrlDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "delete JWKS URL", map[string]interface{}{"id": d.Id()})
	client := meta.(*neon.Client)
	resp, err := client.DeleteProjectJWKS(d.Get("project_id").(string), d.Id())
	if err == nil {
		err = updateStateJwksUrl(d, resp, nil)
	}
	if err == nil {
		d.SetId(resp.ID)
	}
	tflog.Trace(ctx, "successfully deleted JWKS URL", map[string]interface{}{"id": d.Id()})
	return err
}

func resourceJwksUrlCreateRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceJwksUrlCreate, ctx, d, meta)
}

func resourceJwksUrlReadRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceJwksUrlRead, ctx, d, meta)
}

func resourceJwksUrlDeleteRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceJwksUrlDelete, ctx, d, meta)
}

func resourceJwksUrlImport(_ context.Context, _ *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	return nil, errors.New("the resource does not support import, please recreate it instead")
}
