package provider

import (
	"context"
	"errors"
	"strconv"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func newSchemaQuota() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Computed: true,
		Description: `Per-project consumption quota. If the quota is exceeded, all active computes
are automatically suspended and it will not be possible to start them with
an API method call or incoming proxy connections. The only exception is
logical_size_bytes, which is applied on per-branch basis, i.e., only the
compute on the branch that exceeds the logical_size quota will be suspended.

Quotas are enforced based on per-project consumption metrics with the same names,
which are reset at the end of each billing period (the first day of the month).
Logical size is also an exception in this case, as it represents the total size
of data stored in a branch, so it is not reset.

The zero value per attributed means 'unlimited'.`,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"active_time_seconds": {
					Type:        schema.TypeInt,
					Optional:    true,
					Computed:    true,
					Description: `The total amount of wall-clock time allowed to be spent by the project's compute endpoints.`,
				},
				"compute_time_seconds": {
					Type:        schema.TypeInt,
					Optional:    true,
					Computed:    true,
					Description: `The total amount of CPU seconds allowed to be spent by the project's compute endpoints.`,
				},
				"written_data_bytes": {
					Type:        schema.TypeInt,
					Optional:    true,
					Computed:    true,
					Description: `Total amount of data written to all of a project's branches.`,
				},
				"data_transfer_bytes": {
					Type:        schema.TypeInt,
					Optional:    true,
					Computed:    true,
					Description: `Total amount of data transferred from all of a project's branches using the proxy.`,
				},
				"logical_size_bytes": {
					Type:        schema.TypeInt,
					Optional:    true,
					Computed:    true,
					Description: `Limit on the logical size of every project's branch.`,
				},
			},
		},
	}
}

func expandSchemaProjectQuota(v []interface{}) neon.ProjectQuota {
	if len(v) == 0 || v[0] == nil {
		return neon.ProjectQuota{}
	}

	mConf := v[0].(map[string]interface{})

	o := neon.ProjectQuota{}

	if v, ok := mConf["active_time_seconds"].(int); ok && v > 0 {
		o.ActiveTimeSeconds = int64(v)
	}

	if v, ok := mConf["compute_time_seconds"].(int); ok && v > 0 {
		o.ComputeTimeSeconds = int64(v)
	}

	if v, ok := mConf["written_data_bytes"].(int); ok && v > 0 {
		o.WrittenDataBytes = int64(v)
	}

	if v, ok := mConf["data_transfer_bytes"].(int); ok && v > 0 {
		o.DataTransferBytes = int64(v)
	}

	if v, ok := mConf["logical_size_bytes"].(int); ok && v > 0 {
		o.LogicalSizeBytes = int64(v)
	}

	return o
}

func newSchemaDefaultEndpointSettings() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Computed: true,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
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
		},
	}
}

func expandSchemaProjectDefaultEndpointSettings(v []interface{}) *neon.DefaultEndpointSettings {
	if v == nil || len(v) == 0 {
		return nil
	}

	mConf := v[0].(map[string]interface{})

	o := &neon.DefaultEndpointSettings{}
	if v, ok := mConf["autoscaling_limit_min_cu"].(float64); ok && v > 0 {
		o.AutoscalingLimitMinCu = neon.ComputeUnit(v)
	}

	if v, ok := mConf["autoscaling_limit_max_cu"].(float64); ok && v > 0 {
		o.AutoscalingLimitMaxCu = neon.ComputeUnit(v)
	}

	if v, ok := mConf["suspend_timeout_seconds"].(int); ok && v > 0 {
		o.SuspendTimeoutSeconds = neon.SuspendTimeoutSeconds(v)
	}

	return o
}

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Description: `Neon Project. 

See details: https://neon.tech/docs/get-started-with-neon/setting-up-a-project/
API: https://api-docs.neon.tech/reference/createproject`,
		SchemaVersion: versionSchema,
		Importer: &schema.ResourceImporter{
			StateContext: resourceProjectImport,
		},
		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		UpdateContext: resourceProjectUpdateRetry,
		DeleteContext: resourceProjectDeleteRetry,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Project ID.",
			},
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
				ValidateFunc: func(i interface{}, _ string) (warns []string, errs []error) {
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
			"store_password": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether or not passwords are stored for roles in the Neon project. Storing passwords facilitates access to Neon features that require authorization.",
			},
			"history_retention_seconds": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The number of seconds to retain the point-in-time restore (PITR) backup history for this project",
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
			"quota":                     newSchemaQuota(),
			"default_endpoint_settings": newSchemaDefaultEndpointSettings(),
			"branch":                    newBranchSchema(),
			// computed fields
			"default_branch_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default branch ID.",
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

func newBranchSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Computed: true,
		ForceNew: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Branch ID.",
				},
				"name": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					ForceNew:    true,
					Description: "The branch name. If not specified, the default branch name will be used.",
				},
				"role_name": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					ForceNew:    true,
					Description: "The role name. If not specified, the default role name will be used.",
				},
				"database_name": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					ForceNew:    true,
					Description: "The database name. If not specified, the default database name will be used.",
				},
			},
		},
	}
}

func expandSchemaCreateProjectBranch(v []interface{}) *neon.ProjectCreateRequestProjectBranch {
	if v == nil || len(v) == 0 {
		return nil
	}

	mConf := v[0].(map[string]interface{})

	o := &neon.ProjectCreateRequestProjectBranch{}
	if v, ok := mConf["name"].(string); ok && v != "" {
		o.Name = &v
	}

	if v, ok := mConf["role_name"].(string); ok && v != "" {
		o.RoleName = &v
	}

	if v, ok := mConf["database_name"].(string); ok && v != "" {
		o.RoleName = &v
	}

	return o
}

func updateStateProject(d *schema.ResourceData, r neon.Project) error {
	if err := d.Set("name", r.Name); err != nil {
		return err
	}
	if err := d.Set("region_id", r.RegionID); err != nil {
		return err
	}
	if err := d.Set("pg_version", r.PgVersion); err != nil {
		return err
	}
	if err := d.Set("history_retention_seconds", r.HistoryRetentionSeconds); err != nil {
		return err
	}
	if err := d.Set("compute_provisioner", string(r.Provisioner)); err != nil {
		return err
	}
	if err := d.Set("store_password", r.StorePasswords); err != nil {
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

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Trace(ctx, "created Project")

	projectDef := neon.ProjectCreateRequestProject{
		Name:           pointer(d.Get("name").(string)),
		Provisioner:    pointer(neon.Provisioner(d.Get("compute_provisioner").(string))),
		RegionID:       pointer(d.Get("region_id").(string)),
		StorePasswords: pointer(d.Get("store_password").(bool)),
	}
	if v, ok := d.Get("history_retention_seconds").(int); ok && v > 0 {
		projectDef.HistoryRetentionSeconds = pointer(int64(v))
	}

	if v, ok := d.Get("pg_version").(int); ok && v > 0 {
		projectDef.PgVersion = pointer(neon.PgVersion(v))
	}

	if v, ok := d.Get("quota").([]interface{}); ok && len(v) > 0 && v[0] != nil {
		projectDef.Settings = &neon.ProjectSettingsData{Quota: expandSchemaProjectQuota(v)}
	}

	if v, ok := d.Get("branch").([]interface{}); ok && len(v) > 0 && v[0] != nil {
		projectDef.Branch = expandSchemaCreateProjectBranch(v)
	}

	if v, ok := d.Get("default_endpoint_settings").([]interface{}); ok && len(v) > 0 && v[0] != nil {
		projectDef.DefaultEndpointSettings = expandSchemaProjectDefaultEndpointSettings(v)
	}

	client := meta.(neon.Client)

	resp, err := client.CreateProject(
		neon.ProjectCreateRequest{
			Project: projectDef,
		},
	)

	if err != nil {
		return diag.FromErr(err)
	}

	project := resp.ProjectResponse.Project

	d.SetId(project.ID)

	if err := updateStateProject(d, project); err != nil {
		return diag.FromErr(err)
	}

	quota := project.Settings.Quota
	if err := d.Set(
		"quota", []interface{}{
			map[string]interface{}{
				"active_time_seconds":  int(quota.ActiveTimeSeconds),
				"compute_time_seconds": int(quota.ComputeTimeSeconds),
				"written_data_bytes":   int(quota.WrittenDataBytes),
				"data_transfer_bytes":  int(quota.DataTransferBytes),
				"logical_size_bytes":   int(quota.LogicalSizeBytes),
			},
		},
	); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(
		"default_endpoint_settings", []interface{}{
			map[string]interface{}{
				"autoscaling_limit_min_cu": float64(project.DefaultEndpointSettings.AutoscalingLimitMinCu),
				"autoscaling_limit_max_cu": float64(project.DefaultEndpointSettings.AutoscalingLimitMaxCu),
				"suspend_timeout_seconds":  int(project.DefaultEndpointSettings.SuspendTimeoutSeconds),
			},
		},
	); err != nil {
		return diag.FromErr(err)
	}

	if err := updateDefaultBranchDBEndpoint(
		resp.BranchResponse.Branch, resp.EndpointsResponse.Endpoints, resp.DatabasesResponse.Databases,
		resp.RolesResponse.Roles, d,
	); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func updateDefaultBranchDBEndpoint(
	mainBranch neon.Branch, endpoints []neon.Endpoint, databases []neon.Database, roles []neon.Role,
	d *schema.ResourceData,
) error {
	var mainEndpoint neon.Endpoint
	for _, el := range endpoints {
		if !el.Disabled && el.BranchID == mainBranch.ID {
			mainEndpoint = el
			break
		}
	}

	var mainDB neon.Database
	for _, el := range databases {
		if el.BranchID == mainBranch.ID {
			mainDB = el
			break
		}
	}

	var mainRole neon.Role
	for _, el := range roles {
		if el.BranchID == mainBranch.ID && el.Name == mainDB.OwnerName {
			mainRole = el
			if el.Protected {
				mainRole = el
				break
			}
		}
	}

	if err := d.Set(
		"branch", []interface{}{
			map[string]interface{}{
				"id":            mainBranch.ID,
				"name":          mainBranch.Name,
				"role_name":     mainRole.Name,
				"database_name": mainDB.Name,
			},
		},
	); err != nil {
		return err
	}

	if err := d.Set("default_branch_id", mainBranch.ID); err != nil {
		return err
	}

	if err := d.Set("database_password", mainRole.Password); err != nil {
		return err
	}

	connectionURI := "postgres://" + mainRole.Name + ":" + mainRole.Password + "@" + mainEndpoint.Host +
		"/" + mainDB.Name
	if err := d.Set("connection_uri", connectionURI); err != nil {
		return err
	}

	if err := d.Set("database_host", mainEndpoint.Host); err != nil {
		return err
	}
	if err := d.Set("database_name", mainDB.Name); err != nil {
		return err
	}
	if err := d.Set("database_user", mainRole.Name); err != nil {
		return err
	}

	return nil
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "update Project")

	if !d.HasChanges("name", "history_retention_seconds", "default_endpoint_settings", "quota") {
		return nil
	}

	req := neon.ProjectUpdateRequestProject{
		HistoryRetentionSeconds: pointer(int64(d.Get("history_retention_seconds").(int))),
		Name:                    pointer(d.Get("name").(string)),
	}
	req.Settings.Quota = expandSchemaProjectQuota(d.Get("quota").([]interface{}))
	req.DefaultEndpointSettings = expandSchemaProjectDefaultEndpointSettings(
		d.Get("default_endpoint_settings").([]interface{}),
	)

	_, err := meta.(neon.Client).UpdateProject(d.Id(), neon.ProjectUpdateRequest{Project: req})

	return err
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Trace(ctx, "get Project")

	client := meta.(neon.Client)

	resp, err := client.GetProject(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	project := resp.Project

	if err := updateStateProject(d, project); err != nil {
		return diag.FromErr(err)
	}

	quota := project.Settings.Quota
	if err := d.Set(
		"quota", []interface{}{
			map[string]interface{}{
				"active_time_seconds":  int(quota.ActiveTimeSeconds),
				"compute_time_seconds": int(quota.ComputeTimeSeconds),
				"written_data_bytes":   int(quota.WrittenDataBytes),
				"data_transfer_bytes":  int(quota.DataTransferBytes),
				"logical_size_bytes":   int(quota.LogicalSizeBytes),
			},
		},
	); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(
		"default_endpoint_settings", []interface{}{
			map[string]interface{}{
				"autoscaling_limit_min_cu": float64(project.DefaultEndpointSettings.AutoscalingLimitMinCu),
				"autoscaling_limit_max_cu": float64(project.DefaultEndpointSettings.AutoscalingLimitMaxCu),
				"suspend_timeout_seconds":  int(project.DefaultEndpointSettings.SuspendTimeoutSeconds),
			},
		},
	); err != nil {
		return diag.FromErr(err)
	}

	branches, err := client.ListProjectBranches(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var branchMain neon.Branch
	for _, v := range branches.Branches {
		if v.Primary {
			branchMain = v
			break
		}
	}
	if branchMain.ID == "" {
		return nil
	}

	endpoints, err := client.ListProjectBranchEndpoints(d.Id(), branchMain.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	roles, err := client.ListProjectBranchRoles(d.Id(), branchMain.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	dbs, err := client.ListProjectBranchDatabases(d.Id(), branchMain.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := updateDefaultBranchDBEndpoint(
		branchMain, endpoints.Endpoints, dbs.Databases, roles.Roles, d,
	); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "delete Project")

	if _, err := meta.(neon.Client).DeleteProject(d.Id()); err != nil {
		return err
	}

	d.SetId("")
	if err := updateStateProject(d, neon.Project{}); err != nil {
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
	if err := d.Set("default_branch_id", ""); err != nil {
		return err
	}
	for _, k := range []string{
		"quota",
		"branch",
		"default_endpoint_settings",
	} {
		if err := d.Set(k, nil); err != nil {
			return err
		}
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
