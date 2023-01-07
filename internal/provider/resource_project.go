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
			StateContext: resourceProjectImport,
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
				Description: "Project name.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
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
				ForceNew:    true,
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
				Computed: true,
			},
			"autoscaling_limit_max_cu": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Project ID.",
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
			"main_branch_main_endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint to access database",
			},
			"main_branch_main_role_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Initial role of the API key owner.",
			},
		},
	}
}

func updateStateProject(d *schema.ResourceData, r neon.ProjectResponse) error {
	if err := d.Set("name", r.Project.Name); err != nil {
		return err
	}
	if err := d.Set("region_id", r.Project.RegionID); err != nil {
		return err
	}
	if err := d.Set("pg_version", int(r.Project.PgVersion)); err != nil {
		return err
	}
	if err := d.Set("pg_settings", pgSettingsToMap(r.Project.DefaultEndpointSettings.PgSettings)); err != nil {
		return err
	}
	if err := d.Set("cpu_quota_sec", int(r.Project.DefaultEndpointSettings.Quota.CpuQuotaSec)); err != nil {
		return err
	}
	if err := d.Set("branch_logical_size_limit", int(r.Project.BranchLogicalSizeLimit)); err != nil {
		return err
	}
	if err := d.Set("created_at", r.Project.CreatedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("updated_at", r.Project.UpdatedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	return nil
}

func setMainBranchInfo(d *schema.ResourceData, client neon.Client) diag.Diagnostics {
	br, err := client.ListProjectBranches(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	for _, branch := range br.Branches {
		if branch.Name == "main" {
			r, err := client.ListProjectBranchRoles(d.Id(), branch.ID)
			if err != nil {
				return diag.FromErr(err)
			}

			for _, role := range r.Roles {
				if role.Name == "web_access" {
					continue
				}
				_ = d.Set("main_branch_main_role_name", role.Name)
				break
			}
			break
		}
	}

	endpoints, err := client.ListProjectEndpoints(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("main_branch_main_endpoint", endpoints.Endpoints[0].Host)

	return nil
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Trace(ctx, "created Project")

	client := meta.(neon.Client)
	resp, err := client.CreateProject(
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
	if err := updateStateProject(d, resp.ProjectResponse); err != nil {
		return diag.FromErr(err)
	}

	return setMainBranchInfo(d, client)
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Trace(ctx, "update Project")

	resp, err := meta.(neon.Client).UpdateProject(
		d.Id(),
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

	return diag.FromErr(updateStateProject(d, resp.ProjectResponse))
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Trace(ctx, "get Project")

	client := meta.(neon.Client)

	resp, err := client.GetProject(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := updateStateProject(d, resp); err != nil {
		return diag.FromErr(err)
	}

	return setMainBranchInfo(d, client)
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Trace(ctx, "delete Project")

	if _, err := meta.(neon.Client).DeleteProject(d.Id()); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	if err := updateStateProject(d, neon.ProjectResponse{}); err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("main_branch_main_endpoint", "")
	_ = d.Set("main_branch_main_role_name", "")

	return nil
}

func resourceProjectImport(ctx context.Context, d *schema.ResourceData, meta interface{}) (
	[]*schema.ResourceData, error,
) {
	if diags := resourceProjectRead(ctx, d, meta); diags.HasError() {
		return nil, errors.New(diags[0].Summary)
	}
	return []*schema.ResourceData{d}, nil
}
