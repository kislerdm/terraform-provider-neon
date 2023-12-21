package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

const providerDefaultHistoryRetentionSeconds = int(time.Hour/time.Second) * 24 * 7

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Description: `Neon Project. 

See details: https://neon.tech/docs/get-started-with-neon/setting-up-a-project/
API: https://api-docs.neon.tech/reference/createproject`,
		SchemaVersion: 8,
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
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      providerDefaultHistoryRetentionSeconds,
				ValidateFunc: intValidationNotNegative,
				Description: `The number of seconds to retain the point-in-time restore (PITR) backup history for this project. 
Default: 7 days, see https://neon.tech/docs/reference/glossary#point-in-time-restore.`,
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
			"quota":                     schemaQuota,
			"default_endpoint_settings": schemaDefaultEndpointSettings,
			"branch":                    schemaDefaultBranch,
			"allowed_ips": {
				Type:     schema.TypeList,
				MinItems: 1,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Description: `A list of IP addresses that are allowed to connect to the endpoints. 
Note that the feature is available to the Neon Pro Plan only. Details: https://neon.tech/docs/manage/projects#configure-ip-allow`,
			},
			"allowed_ips_primary_branch_only": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: `Apply the allow-list to the primary branch only.
Note that the feature is available to the Neon Pro Plan only.`,
			},
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

var schemaQuota = &schema.Schema{
	Type:     schema.TypeList,
	MaxItems: 1,
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
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: intValidationNotNegative,
				Description:  `The total amount of wall-clock time allowed to be spent by the project's compute endpoints.`,
			},
			"compute_time_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: intValidationNotNegative,
				Description:  `The total amount of CPU seconds allowed to be spent by the project's compute endpoints.`,
			},
			"written_data_bytes": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: intValidationNotNegative,
				Description:  `Total amount of data written to all of a project's branches.`,
			},
			"data_transfer_bytes": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: intValidationNotNegative,
				Description:  `Total amount of data transferred from all of a project's branches using the proxy.`,
			},
			"logical_size_bytes": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: intValidationNotNegative,
				Description:  `Limit on the logical size of every project's branch.`,
			},
		},
	},
}

func mapToQuotaSettings(v map[string]interface{}) *neon.ProjectQuota {
	o := neon.ProjectQuota{}

	if v, ok := v["active_time_seconds"].(int); ok && v > 0 {
		o.ActiveTimeSeconds = pointer(int64(v))
	}

	if v, ok := v["compute_time_seconds"].(int); ok && v > 0 {
		o.ComputeTimeSeconds = pointer(int64(v))
	}

	if v, ok := v["written_data_bytes"].(int); ok && v > 0 {
		o.WrittenDataBytes = pointer(int64(v))
	}

	if v, ok := v["data_transfer_bytes"].(int); ok && v > 0 {
		o.DataTransferBytes = pointer(int64(v))
	}

	if v, ok := v["logical_size_bytes"].(int); ok && v > 0 {
		o.LogicalSizeBytes = pointer(int64(v))
	}

	return &o
}

var schemaDefaultEndpointSettings = &schema.Schema{
	Type:     schema.TypeList,
	MaxItems: 1,
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
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: intValidationNotNegative,
				Description: `Duration of inactivity in seconds after which the compute endpoint is automatically suspended. 
The value 0 means use the global default.
The value -1 means never suspend. The default value is 300 seconds (5 minutes).
The maximum value is 604800 seconds (1 week)`,
			},
		},
	},
}

func mapToDefaultEndpointsSettings(v map[string]interface{}) *neon.DefaultEndpointSettings {
	o := neon.DefaultEndpointSettings{}
	if v, ok := v["autoscaling_limit_min_cu"].(float64); ok && v > 0 {
		o.AutoscalingLimitMinCu = pointer(neon.ComputeUnit(v))
	}

	if v, ok := v["autoscaling_limit_max_cu"].(float64); ok && v > 0 {
		o.AutoscalingLimitMaxCu = pointer(neon.ComputeUnit(v))
	}

	if v, ok := v["suspend_timeout_seconds"].(int); ok && v > 0 {
		o.SuspendTimeoutSeconds = pointer(neon.SuspendTimeoutSeconds(v))
	}
	return &o
}

var schemaDefaultBranch = &schema.Schema{
	Type:     schema.TypeList,
	MaxItems: 1,
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

func mapToBranchSettings(v map[string]interface{}) *neon.ProjectCreateRequestProjectBranch {
	o := &neon.ProjectCreateRequestProjectBranch{}

	if v, ok := v["name"].(string); ok && v != "" {
		o.Name = pointer(v)
	}

	if v, ok := v["database_name"].(string); ok && v != "" {
		o.DatabaseName = pointer(v)
	}

	if v, ok := v["role_name"].(string); ok && v != "" {
		o.RoleName = pointer(v)
	}

	return o
}

func newDbConnectionInfo(
	c sdkProject,
	projectID string,
	branchID string,
	endpoints []neon.Endpoint,
	databases []neon.Database,
) (dbConnectionInfo, error) {
	o := dbConnectionInfo{}

	for _, el := range endpoints {
		if !el.Disabled && el.BranchID == branchID {
			o.host = el.Host
			break
		}
	}

	for _, el := range databases {
		if el.BranchID == branchID {
			o.dbName = el.Name
			o.userName = el.OwnerName
			break
		}
	}

	resp, err := c.GetProjectBranchRolePassword(projectID, branchID, o.userName)
	if err != nil {
		return dbConnectionInfo{}, err
	}

	o.pass = resp.Password

	return o, nil
}

type dbConnectionInfo struct {
	userName string
	pass     string
	dbName   string
	host     string
}

func (i dbConnectionInfo) connectionURI() string {
	if i.userName == "" || i.host == "" || i.dbName == "" || i.pass == "" {
		return ""
	}
	return "postgres://" + i.userName + ":" + i.pass + "@" + i.host + "/" + i.dbName
}

func updateStateProject(
	d *schema.ResourceData, r neon.Project,
	defaultBranchID, defaultBranchName string,
	dbConnectionInfo dbConnectionInfo,
) error {
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

	if r.DefaultEndpointSettings != nil {
		defaultEndpointSettings := map[string]interface{}{}
		if r.DefaultEndpointSettings.AutoscalingLimitMinCu != nil {
			defaultEndpointSettings["autoscaling_limit_min_cu"] = float64(*r.DefaultEndpointSettings.AutoscalingLimitMinCu)
		}
		if r.DefaultEndpointSettings.AutoscalingLimitMaxCu != nil {
			defaultEndpointSettings["autoscaling_limit_max_cu"] = float64(*r.DefaultEndpointSettings.AutoscalingLimitMaxCu)
		}
		if r.DefaultEndpointSettings.SuspendTimeoutSeconds != nil {
			defaultEndpointSettings["suspend_timeout_seconds"] = float64(*r.DefaultEndpointSettings.SuspendTimeoutSeconds)
		}
		if err := d.Set("default_endpoint_settings", []interface{}{defaultEndpointSettings}); err != nil {
			return err
		}
	}

	if err := d.Set(
		"branch", []interface{}{
			map[string]interface{}{
				"id":            defaultBranchID,
				"name":          defaultBranchName,
				"role_name":     dbConnectionInfo.userName,
				"database_name": dbConnectionInfo.dbName,
			},
		},
	); err != nil {
		return err
	}

	if r.Settings != nil {
		if r.Settings.Quota != nil {
			if err := d.Set(
				"quota", []interface{}{
					map[string]interface{}{
						"active_time_seconds":  int(*r.Settings.Quota.ActiveTimeSeconds),
						"compute_time_seconds": int(*r.Settings.Quota.ComputeTimeSeconds),
						"written_data_bytes":   int(*r.Settings.Quota.WrittenDataBytes),
						"data_transfer_bytes":  int(*r.Settings.Quota.DataTransferBytes),
						"logical_size_bytes":   int(*r.Settings.Quota.LogicalSizeBytes),
					},
				},
			); err != nil {
				return err
			}
		}

		if r.Settings.AllowedIps != nil {
			if err := d.Set("allowed_ips", r.Settings.AllowedIps.Ips); err != nil {
				return err
			}
			if err := d.Set("allowed_ips_primary_branch_only", r.Settings.AllowedIps.PrimaryBranchOnly); err != nil {
				return err
			}
		}
	}

	if err := d.Set("default_branch_id", defaultBranchID); err != nil {
		return err
	}

	if err := d.Set("database_host", dbConnectionInfo.host); err != nil {
		return err
	}

	if err := d.Set("database_name", dbConnectionInfo.dbName); err != nil {
		return err
	}

	if err := d.Set("database_user", dbConnectionInfo.userName); err != nil {
		return err
	}

	if err := d.Set("database_password", dbConnectionInfo.pass); err != nil {
		return err
	}

	if err := d.Set("connection_uri", dbConnectionInfo.connectionURI()); err != nil {
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

	if v, ok := d.Get("history_retention_seconds").(int); ok && v >= 0 {
		projectDef.HistoryRetentionSeconds = pointer(int64(v))
	}

	if v, ok := d.GetOk("pg_version"); ok && v.(int) > 0 {
		projectDef.PgVersion = pointer(neon.PgVersion(v.(int)))
	}

	if v, ok := d.GetOk("default_endpoint_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if v, ok := v.([]interface{})[0].(map[string]interface{}); ok && len(v) > 0 {
			projectDef.DefaultEndpointSettings = mapToDefaultEndpointsSettings(v)
		}
	}

	if v, ok := d.GetOk("quota"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if v, ok := v.([]interface{})[0].(map[string]interface{}); ok && len(v) > 0 {
			projectDef.Settings = &neon.ProjectSettingsData{
				Quota: mapToQuotaSettings(v),
			}
		}
	}

	if v, ok := d.GetOk("allowed_ips"); ok && len(v.([]interface{})) > 0 {
		var ips = make([]string, len(v.([]interface{})))
		for i, vv := range v.([]interface{}) {
			ips[i] = fmt.Sprintf("%v", vv)
		}

		var q *neon.ProjectQuota
		if projectDef.Settings != nil && projectDef.Settings.Quota != nil {
			q = projectDef.Settings.Quota
		}
		projectDef.Settings = &neon.ProjectSettingsData{
			AllowedIps: &neon.AllowedIps{
				Ips:               ips,
				PrimaryBranchOnly: d.Get("allowed_ips_primary_branch_only").(bool),
			},
			Quota: q,
		}
	}

	if v, ok := d.GetOk("branch"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if v, ok := v.([]interface{})[0].(map[string]interface{}); ok && len(v) > 0 {
			projectDef.Branch = mapToBranchSettings(v)
		}
	}

	client := meta.(sdkProject)

	resp, err := client.CreateProject(
		neon.ProjectCreateRequest{
			Project: projectDef,
		},
	)

	if err != nil {
		return diag.FromErr(err)
	}

	projectID := resp.ProjectResponse.Project.ID
	d.SetId(projectID)

	branch := resp.BranchResponse.Branch
	info, err := newDbConnectionInfo(client, projectID, branch.ID, resp.EndpointsResponse.Endpoints,
		resp.DatabasesResponse.Databases)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := updateStateProject(d, resp.ProjectResponse.Project, branch.ID, branch.Name, info); err != nil {
		return diag.FromErr(err)
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

	if v, ok := d.GetOk("default_endpoint_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if v, ok := v.([]interface{})[0].(map[string]interface{}); ok && len(v) > 0 {
			req.DefaultEndpointSettings = mapToDefaultEndpointsSettings(v)
		}
	}

	if v, ok := d.GetOk("quota"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if v, ok := v.([]interface{})[0].(map[string]interface{}); ok && len(v) > 0 {
			req.Settings = &neon.ProjectSettingsData{
				Quota: mapToQuotaSettings(v),
			}
		}
	}

	if v, ok := d.GetOk("allowed_ips"); ok && len(v.([]interface{})) > 0 {
		var ips = make([]string, len(v.([]interface{})))
		for i, vv := range v.([]interface{}) {
			ips[i] = fmt.Sprintf("%v", vv)
		}
		req.Settings = &neon.ProjectSettingsData{
			Quota: req.Settings.Quota,
			AllowedIps: &neon.AllowedIps{
				Ips:               ips,
				PrimaryBranchOnly: d.Get("allowed_ips_primary_branch_only").(bool),
			},
		}
	}

	_, err := meta.(sdkProject).UpdateProject(d.Id(), neon.ProjectUpdateRequest{Project: req})

	return err
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Trace(ctx, "get Project")

	client := meta.(sdkProject)

	resp, err := client.GetProject(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	project := resp.Project

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
		return diag.FromErr(updateStateProject(d, project, branchMain.ID, branchMain.Name, dbConnectionInfo{}))
	}

	endpoints, err := client.ListProjectBranchEndpoints(d.Id(), branchMain.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	dbs, err := client.ListProjectBranchDatabases(d.Id(), branchMain.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	info, err := newDbConnectionInfo(client, project.ID, branchMain.ID, endpoints.Endpoints, dbs.Databases)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(updateStateProject(d, project, branchMain.ID, branchMain.Name, info))
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "delete Project")

	if _, err := meta.(sdkProject).DeleteProject(d.Id()); err != nil {
		return err
	}

	d.SetId("")
	return updateStateProject(d, neon.Project{}, "", "", dbConnectionInfo{})
}

func resourceProjectImport(ctx context.Context, d *schema.ResourceData, meta interface{}) (
	[]*schema.ResourceData, error,
) {
	if diags := resourceProjectRead(ctx, d, meta); diags.HasError() {
		return nil, errors.New(diags[0].Summary)
	}
	return []*schema.ResourceData{d}, nil
}

type sdkProject interface {
	GetProjectBranchRolePassword(string, string, string) (neon.RolePasswordResponse, error)
	CreateProject(neon.ProjectCreateRequest) (neon.CreatedProject, error)
	UpdateProject(string, neon.ProjectUpdateRequest) (neon.UpdateProjectRespObj, error)
	GetProject(string) (neon.ProjectResponse, error)
	ListProjectBranches(string) (neon.BranchesResponse, error)
	ListProjectBranchEndpoints(string, string) (neon.EndpointsResponse, error)
	DeleteProject(string) (neon.ProjectResponse, error)
	ListProjectBranchDatabases(string, string) (neon.DatabasesResponse, error)
}
