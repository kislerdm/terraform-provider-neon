package provider

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Description:   "Neon Project. See details: https://neon.tech/docs/get-started-with-neon/setting-up-a-project/",
		SchemaVersion: versionSchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		UpdateContext: resourceProjectUpdate,
		DeleteContext: resourceProjectDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Project name.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "AWS Region.",
				ValidateFunc: func(i interface{}, s string) (warns []string, errs []error) {
					switch v := i.(string); v {
					case "aws-us-east-2", "aws-us-west-2", "aws-eu-central-1", "aws-ap-southeast-1":
						return
					default:
						errs = append(
							errs,
							errors.New(
								"region "+v+" is not supported yet: https://neon.tech/docs/introduction/regions/",
							),
						)
						return
					}
				},
			},
			"pg_version": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Postgres version",
				ValidateFunc: func(i interface{}, s string) (warns []string, errs []error) {
					switch v := i.(int); v {
					case 14, 15:
						return
					default:
						errs = append(
							errs, errors.New("postgres version "+strconv.Itoa(v)+" is not supported yet"),
						)
						return
					}
				},
			},
			"pg_settings": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
			},
			"cpu_quota_sec": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Total amount of CPU seconds that is allowed to be spent by the endpoints of that project.",
			},
			"autoscaling_limit_min_cu": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"autoscaling_limit_max_cu": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Project ID.",
			},
			"platform_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Platform type id.",
			},
			"maintenance_starts_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "If set, means project will be in maintenance since that time.",
			},
			"locked": {
				Type:     schema.TypeBool,
				Computed: true,
				Description: `Currently, a project may not have more than one running operations chain.
If there are any running operations, 'locked' will be set to 'true'.
This attributed is considered to be temporary, and could be gone soon.`,
			},
			"proxy_host": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpu_used_sec": {
				Type:     schema.TypeInt,
				Computed: true,
				Description: `CPU seconds used by all the endpoints of the project, including deleted ones.
This value is reset at the beginning of each billing period.
Examples:
1. Having endpoint used 1 CPU for 1 sec, that's cpu_used_sec=1.
2. Having endpoint used 2 CPU simultaneously for 1 sec, that's cpu_used_sec=2.`,
			},
			"branch_logical_size_limit": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Project creation timestamp.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Project last update timestamp.",
			},
		},
	}
}

func pgSettingsToMap(v neon.PgSettingsData) map[string]interface{} {
	o := make(map[string]interface{}, len(v))
	for k, v := range v {
		o[k] = v
	}
	return o
}

func mapToPgSettings(v map[string]interface{}) neon.PgSettingsData {
	o := make(neon.PgSettingsData, len(v))
	for k, v := range v {
		o[k] = v
	}
	return o
}

func updateStateProject(d *schema.ResourceData, r neon.ProjectResponse) {
	_ = d.Set("name", r.Project.Name)
	_ = d.Set("region_id", r.Project.RegionID)
	_ = d.Set("pg_version", int(r.Project.PgVersion))
	_ = d.Set("pg_settings", pgSettingsToMap(r.Project.DefaultEndpointSettings.PgSettings))
	_ = d.Set("cpu_quota_sec", int(r.Project.DefaultEndpointSettings.Quota.CpuQuotaSec))
	_ = d.Set("platform_id", r.Project.PlatformID)
	_ = d.Set("maintenance_starts_at", r.Project.MaintenanceStartsAt.Format(time.RFC3339))
	_ = d.Set("locked", r.Project.Locked)
	_ = d.Set("proxy_host", r.Project.ProxyHost)
	_ = d.Set("cpu_used_sec", int(r.Project.CpuUsedSec))
	_ = d.Set("branch_logical_size_limit", int(r.Project.BranchLogicalSizeLimit))
	_ = d.Set("created_at", r.Project.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", r.Project.UpdatedAt.Format(time.RFC3339))
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Trace(ctx, "created Project")

	resp, err := meta.(neon.Client).CreateProject(
		neon.ProjectCreateRequest{
			Project: neon.ProjectCreateRequestProject{
				AutoscalingLimitMinCu:   int32(d.Get("autoscaling_limit_min_cu").(int)),
				AutoscalingLimitMaxCu:   int32(d.Get("autoscaling_limit_max_cu").(int)),
				RegionID:                d.Get("region_id").(string),
				DefaultEndpointSettings: mapToPgSettings(d.Get("pg_settings").(map[string]interface{})),
				PgVersion:               neon.PgVersion(d.Get("pg_version").(int)),
				Quota:                   neon.ProjectQuota{CpuQuotaSec: int64(d.Get("cpu_quota_sec").(int))},
				Name:                    d.Get("name").(string),
			},
		},
	)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ProjectResponse.Project.ID)
	updateStateProject(d, resp.ProjectResponse)
	return nil
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Trace(ctx, "update Project")

	for {
		resourceProjectRead(ctx, d, meta)
		if !d.Get("locked").(bool) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	resp, err := meta.(neon.Client).UpdateProject(
		d.Get("id").(string),
		neon.ProjectUpdateRequest{
			Project: neon.ProjectUpdateRequestProject{
				DefaultEndpointSettings: mapToPgSettings(d.Get("pg_settings").(map[string]interface{})),
				Quota: neon.ProjectQuota{
					CpuQuotaSec: int64(d.Get("cpu_quota_sec").(int)),
				},
				AutoscalingLimitMinCu: int32(d.Get("autoscaling_limit_min_cu").(int)),
				AutoscalingLimitMaxCu: int32(d.Get("autoscaling_limit_max_cu").(int)),
				Name:                  d.Get("name").(string),
			},
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	updateStateProject(d, resp.ProjectResponse)
	return nil
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Trace(ctx, "get Project")

	resp, err := meta.(neon.Client).GetProject(d.Get("id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	updateStateProject(d, resp)
	return nil
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Trace(ctx, "delete Project")

	if _, err := meta.(neon.Client).DeleteProject(d.Get("id").(string)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	updateStateProject(d, neon.ProjectResponse{})
	return nil
}
