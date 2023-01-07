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
		UpdateContext: resourceProjectUpdateRetry,
		DeleteContext: resourceProjectDeleteRetry,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Project name.",
			},
			"region_id": schemaRegionID,
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
			"database_host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default database host.",
			},
			"database_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default database name.",
			},
			"database_user": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default database role.",
			},
			"database_password": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Default database access password.",
			},
			"connection_uri": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Default connection uri. **Note** that it contains access credentials.",
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
	if err := d.Set("pg_version", r.Project.PgVersion); err != nil {
		return err
	}
	if err := d.Set("pg_settings", pgSettingsToMap(r.Project.DefaultEndpointSettings.PgSettings)); err != nil {
		return err
	}
	if err := d.Set("cpu_quota_sec", r.Project.DefaultEndpointSettings.Quota.CpuQuotaSec); err != nil {
		return err
	}
	if err := d.Set("branch_logical_size_limit", r.Project.BranchLogicalSizeLimit); err != nil {
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

func resourceProjectDeleteRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceProjectDelete, ctx, d, meta)
}

func resourceProjectUpdateRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceProjectUpdate, ctx, d, meta)
}

func setMainBranchInfo(d *schema.ResourceData, client neon.Client) diag.Diagnostics {
	br, err := client.ListProjectBranches(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	endpoints, err := client.ListProjectEndpoints(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	databaseHost := endpoints.Endpoints[0].Host
	if err := d.Set("database_host", databaseHost); err != nil {
		return diag.FromErr(err)
	}

	for _, branch := range br.Branches {
		if branch.Name == "main" {
			r, err := client.ListProjectBranchRoles(d.Id(), branch.ID)
			if err != nil {
				return diag.FromErr(err)
			}

			if err := setRole(d, r.Roles); err != nil {
				return diag.FromErr(err)
			}

			o, err := client.ListProjectBranchDatabases(d.Id(), branch.ID)
			if err != nil {
				return diag.FromErr(err)
			}

			databaseName := o.Databases[0].Name
			if err := d.Set("database_name", databaseName); err != nil {
				return diag.FromErr(err)
			}

			databaseUser := d.Get("database_user").(string)
			databasePassword := d.Get("database_password").(string)
			if databaseUser != "" && databasePassword != "" {
				connectionURI := "postgres://" + databaseUser + ":" + databasePassword + "@" + databaseHost + "/" + databaseName
				if err := d.Set("connection_uri", connectionURI); err != nil {
					return diag.FromErr(err)
				}
			}

			break
		}
	}

	return nil
}

func setRole(d *schema.ResourceData, r []neon.Role) error {
	for _, role := range r {
		if role.Name == "web_access" {
			continue
		}
		if err := d.Set("database_user", role.Name); err != nil {
			return err
		}
		if err := d.Set("database_password", role.Password); err != nil {
			return err
		}
		break
	}
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

	if err := setRole(d, resp.Roles); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("database_host", resp.Endpoints[0].Host); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("database_name", resp.Databases[0].Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("connection_uri", resp.ConnectionUris[0].ConnectionURI); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
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
		return err
	}

	return updateStateProject(d, resp.ProjectResponse)
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

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "delete Project")

	if _, err := meta.(neon.Client).DeleteProject(d.Id()); err != nil {
		return err
	}

	d.SetId("")
	if err := updateStateProject(d, neon.ProjectResponse{}); err != nil {
		return err
	}

	if err := d.Set("database_name", ""); err != nil {
		return err
	}
	if err := d.Set("database_host", ""); err != nil {
		return err
	}
	if err := d.Set("database_user", ""); err != nil {
		return err
	}
	if err := d.Set("database_password", ""); err != nil {
		return err
	}
	if err := d.Set("connection_uri", ""); err != nil {
		return err
	}
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
