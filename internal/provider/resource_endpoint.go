package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

const endpointTypeRW = "read_write"

func resourceEndpoint() *schema.Resource {
	return &schema.Resource{
		Description: `Project Endpoint. See details: https://neon.tech/docs/manage/endpoints/

!> Note that "read_write" type is only supported yet, and a single "read_write" endpoint is permitted per branch. ` +
			`It means that no additional endpoints can be provisioned for branches with existing endpoints. ` +
			`It also means that no endpoints can be created for branches provisioned with this terraform provider ` +
			`because every branch has the default endpoint attached.`,
		SchemaVersion: 8,
		Importer: &schema.ResourceImporter{
			StateContext: resourceEndpointImport,
		},
		CreateContext: resourceEndpointCreateRetry,
		ReadContext:   resourceEndpointReadRetry,
		UpdateContext: resourceEndpointUpdateRetry,
		DeleteContext: resourceEndpointDeleteRetry,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint ID.",
			},
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
				Optional:    true,
				Default:     endpointTypeRW,
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
				Type:         schema.TypeFloat,
				ValidateFunc: validateAutoscallingLimit,
				Optional:     true,
				Computed:     true,
			},
			"autoscaling_limit_max_cu": {
				Type:         schema.TypeFloat,
				ValidateFunc: validateAutoscallingLimit,
				Optional:     true,
				Computed:     true,
			},
			"pg_settings": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"pooler_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
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
			"compute_provisioner": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: `Provisioner The Neon compute provisioner.
Specify the k8s-neonvm provisioner to create a compute endpoint that supports Autoscaling.
`,
				ValidateFunc: func(i interface{}, s string) (warns []string, errs []error) {
					switch v := i.(string); v {
					case "k8s-pod", "k8s-neonvm":
					default:
						errs = append(
							errs,
							errors.New(
								v+" is not supported for "+s+
									". See details: https://api-docs.neon.tech/reference/createproject",
							),
						)
					}
					return
				},
			},
			"suspend_timeout_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				Description: `Duration of inactivity in seconds after which the compute endpoint is automatically suspended.
The value 0 means use the global default.
The value -1 means never suspend. The default value is 300 seconds (5 minutes).
The maximum value is 604800 seconds (1 week)`,
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
	if err := d.Set("autoscaling_limit_min_cu", float64(v.AutoscalingLimitMinCu)); err != nil {
		return err
	}
	if err := d.Set("autoscaling_limit_max_cu", float64(v.AutoscalingLimitMaxCu)); err != nil {
		return err
	}
	if v.Settings.PgSettings != nil {
		if err := d.Set("pg_settings", pgSettingsToMap(*v.Settings.PgSettings)); err != nil {
			return err
		}
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
	if err := d.Set("compute_provisioner", string(v.Provisioner)); err != nil {
		return err
	}
	if err := d.Set("suspend_timeout_seconds", int64(v.SuspendTimeoutSeconds)); err != nil {
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
		RegionID:              pointer(d.Get("region_id").(string)),
		PoolerEnabled:         pointer(d.Get("pooler_enabled").(bool)),
		PoolerMode:            pointer(neon.EndpointPoolerMode(d.Get("pooler_mode").(string))),
		Disabled:              pointer(d.Get("disabled").(bool)),
		Provisioner:           pointer(neon.Provisioner(d.Get("compute_provisioner").(string))),
		SuspendTimeoutSeconds: pointer(neon.SuspendTimeoutSeconds(d.Get("suspend_timeout_seconds").(int))),
	}

	if v, ok := d.GetOk("autoscaling_limit_min_cu"); ok {
		cfg.AutoscalingLimitMinCu = pointer(neon.ComputeUnit(v.(float64)))
	}

	if v, ok := d.GetOk("autoscaling_limit_max_cu"); ok {
		cfg.AutoscalingLimitMaxCu = pointer(neon.ComputeUnit(v.(float64)))
	}

	if v, ok := d.GetOk("pg_settings"); ok {
		cfg.Settings = &neon.EndpointSettingsData{
			PgSettings: mapToPgSettings(v.(map[string]interface{})),
		}
	}

	resp, err := meta.(*neon.Client).CreateProjectEndpoint(
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

	resp, err := meta.(*neon.Client).GetProjectEndpoint(
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
		PoolerEnabled:         pointer(d.Get("pooler_enabled").(bool)),
		PoolerMode:            pointer(neon.EndpointPoolerMode(d.Get("pooler_mode").(string))),
		Disabled:              pointer(d.Get("disabled").(bool)),
		BranchID:              pointer(d.Get("branch_id").(string)),
		AutoscalingLimitMinCu: pointer(neon.ComputeUnit(d.Get("autoscaling_limit_min_cu").(float64))),
		AutoscalingLimitMaxCu: pointer(neon.ComputeUnit(d.Get("autoscaling_limit_max_cu").(float64))),
		Provisioner:           pointer(neon.Provisioner(d.Get("compute_provisioner").(string))),
		SuspendTimeoutSeconds: pointer(neon.SuspendTimeoutSeconds(d.Get("suspend_timeout_seconds").(int))),
	}

	if v, ok := d.GetOk("pg_settings"); ok {
		cfg.Settings = &neon.EndpointSettingsData{
			PgSettings: mapToPgSettings(v.(map[string]interface{})),
		}
	}

	resp, err := meta.(*neon.Client).UpdateProjectEndpoint(
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

	resp, err := meta.(*neon.Client).ListProjects(nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	for _, project := range resp.Projects {
		r, err := meta.(*neon.Client).ListProjectEndpoints(project.ID)
		if err != nil {
			return nil, err
		}

		for _, endpoint := range r.Endpoints {
			if endpoint.ID == d.Id() {
				if err := d.Set("project_id", project.ID); err != nil {
					return nil, err
				}
				if err := resourceEndpointRead(ctx, d, meta); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			}
		}

	}

	return nil, errors.New("no endpoint " + d.Id() + " found")
}

func resourceEndpointDeleteRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceEndpointDelete, ctx, d, meta)
}

func resourceEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "delete Endpoint")
	if _, err := meta.(*neon.Client).DeleteProjectEndpoint(d.Get("project_id").(string), d.Id()); err != nil {
		return err
	}
	d.SetId("")
	return updateStateEndpoint(d, neon.Endpoint{})
}
