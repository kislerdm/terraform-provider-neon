package provider

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func resourceEndpoint() *schema.Resource {
	return &schema.Resource{
		Description:   "Project Endpoint. See details: https://neon.tech/docs/manage/endpoints/",
		SchemaVersion: versionSchema,
		Importer: &schema.ResourceImporter{
			StateContext: resourceEndpointImport,
		},
		CreateContext: resourceEndpointCreateRetry,
		ReadContext:   resourceEndpointReadRetry,
		UpdateContext: resourceEndpointUpdateRetry,
		DeleteContext: resourceEndpointDeleteRetry,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Project ID.",
			},
			"branch_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Branch ID.",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `Access type. **Note** that "read_write" is the only supported type yet.`,
				ValidateFunc: func(d interface{}, k string) (warn []string, errs []error) {
					switch v := d.(string); v {
					case "read_write":
						return
					case "read_only":
						warn = append(warn, `"read_write" is only supported option yet`)
					default:
						errs = append(errs, errors.New(v+" is not supported value for "+k))
						return
					}
					return
				},
			},
			"host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint URI.",
			},
			"region_id": schemaRegionID,
			"autoscaling_limit_min_cu": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"autoscaling_limit_max_cu": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"pg_settings": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
			},
			"passwordless_access": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Allow passwordless access.",
			},
			"pooler_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: `Activate connection pooling.
See details: https://neon.tech/docs/connect/connection-pooling`,
			},
			"pooler_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: `Mode of connections pooling.
See details: https://neon.tech/docs/connect/connection-pooling`,
			},
			"disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Disable the endpoint.",
			},
			"proxy_host": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"current_state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint state.",
			},
			"pending_state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint pending state.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint creation timestamp.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint last update timestamp.",
			},
		},
	}
}

func updateStateEndpoint(d *schema.ResourceData, v neon.Endpoint) error {
	if err := d.Set("type", v.Type); err != nil {
		return err
	}
	if err := d.Set("host", v.Host); err != nil {
		return err
	}
	if err := d.Set("region_id", v.RegionID); err != nil {
		return err
	}
	if err := d.Set("autoscaling_limit_min_cu", v.AutoscalingLimitMinCu); err != nil {
		return err
	}
	if err := d.Set("autoscaling_limit_max_cu", v.AutoscalingLimitMaxCu); err != nil {
		return err
	}
	if err := d.Set("pg_settings", pgSettingsToMap(v.Settings.PgSettings)); err != nil {
		return err
	}
	if err := d.Set("passwordless_access", v.PasswordlessAccess); err != nil {
		return err
	}
	if err := d.Set("pooler_enabled", v.PoolerEnabled); err != nil {
		return err
	}
	if err := d.Set("pooler_mode", string(v.PoolerMode)); err != nil {
		return err
	}
	if err := d.Set("disabled", v.Disabled); err != nil {
		return err
	}
	if err := d.Set("proxy_host", v.ProxyHost); err != nil {
		return err
	}
	if err := d.Set("current_state", v.CurrentState); err != nil {
		return err
	}
	if err := d.Set("pending_state", v.PendingState); err != nil {
		return err
	}
	if err := d.Set("created_at", v.CreatedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("updated_at", v.UpdatedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	return nil
}

func resourceEndpointCreateRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceEndpointCreate, ctx, d, meta)
}

func resourceEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "created Endpoint")

	cfg := neon.EndpointCreateRequestEndpoint{
		BranchID:              d.Get("branch_id").(string),
		Type:                  neon.EndpointType(d.Get("type").(string)),
		RegionID:              d.Get("region_id").(string),
		PoolerEnabled:         d.Get("pooler_enabled").(bool),
		AutoscalingLimitMinCu: int32(d.Get("autoscaling_limit_min_cu").(int)),
		AutoscalingLimitMaxCu: int32(d.Get("autoscaling_limit_max_cu").(int)),
		PoolerMode:            neon.EndpointPoolerMode(d.Get("pooler_mode").(string)),
		PasswordlessAccess:    d.Get("passwordless_access").(bool),
		Disabled:              d.Get("disabled").(bool),
	}

	if v, ok := d.GetOk("pg_settings"); ok {
		cfg.Settings = &neon.EndpointSettingsData{
			PgSettings: mapToPgSettings(v.(map[string]interface{})),
		}
	}

	resp, err := meta.(neon.Client).CreateProjectEndpoint(
		d.Get("project_id").(string),
		neon.EndpointCreateRequest{Endpoint: cfg},
	)
	if err != nil {
		return err
	}

	d.SetId(resp.Endpoint.ID)
	return updateStateEndpoint(d, resp.EndpointResponse.Endpoint)
}

func resourceEndpointReadRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceEndpointRead, ctx, d, meta)
}

func resourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "read Endpoint")

	resp, err := meta.(neon.Client).GetProjectEndpoint(
		d.Get("project_id").(string),
		d.Id(),
	)
	if err != nil {
		return err
	}

	return updateStateEndpoint(d, resp.Endpoint)
}

func resourceEndpointUpdateRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceEndpointUpdate, ctx, d, meta)
}

func resourceEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "update Endpoint")

	cfg := neon.EndpointUpdateRequestEndpoint{
		PoolerEnabled:         d.Get("pooler_enabled").(bool),
		PoolerMode:            neon.EndpointPoolerMode(d.Get("pooler_mode").(string)),
		Disabled:              d.Get("disabled").(bool),
		PasswordlessAccess:    d.Get("passwordless_access").(bool),
		BranchID:              d.Get("branch_id").(string),
		AutoscalingLimitMinCu: int32(d.Get("autoscaling_limit_min_cu").(int)),
		AutoscalingLimitMaxCu: int32(d.Get("autoscaling_limit_max_cu").(int)),
	}

	if v, ok := d.GetOk("pg_settings"); ok {
		cfg.Settings = &neon.EndpointSettingsData{
			PgSettings: mapToPgSettings(v.(map[string]interface{})),
		}
	}

	resp, err := meta.(neon.Client).UpdateProjectEndpoint(
		d.Get("project_id").(string),
		d.Id(),
		neon.EndpointUpdateRequest{Endpoint: cfg},
	)
	if err != nil {
		return err
	}
	return updateStateEndpoint(d, resp.EndpointResponse.Endpoint)
}

func resourceEndpointImport(ctx context.Context, d *schema.ResourceData, meta interface{}) (
	[]*schema.ResourceData, error,
) {
	tflog.Trace(ctx, "import Endpoint")

	resp, err := meta.(neon.Client).ListProjects()
	if err != nil {
		return nil, err
	}

	for _, project := range resp.Projects {
		if err := d.Set("project_id", project.ID); err != nil {
			return nil, err
		}
		switch err := resourceEndpointRead(ctx, d, meta).(type) {
		case nil:
			return []*schema.ResourceData{d}, nil
		case neon.Error:
			if err.HTTPCode == http.StatusNotFound {
				continue
			}
		default:
			return nil, err
		}
	}
	return nil, errors.New("no endpoint " + d.Id() + " found")
}

func resourceEndpointDeleteRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceEndpointDelete, ctx, d, meta)
}

func resourceEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "delete Endpoint")
	if _, err := meta.(neon.Client).DeleteProjectEndpoint(d.Get("project_id").(string), d.Id()); err != nil {
		return err
	}
	d.SetId("")
	return updateStateEndpoint(d, neon.Endpoint{})
}
