package provider

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
	"github.com/kislerdm/terraform-provider-neon/internal/types"
)

const providerDefaultHistoryRetentionSeconds = int(time.Hour/time.Second) * 24 * 7

func newStoreProjectPasswordDefault() *schema.Schema {
	o := types.NewOptionalTristateBool(`Whether or not passwords are stored for roles in the Neon project. 
Storing passwords facilitates access to Neon features that require authorization.`, false)
	o.Default = types.ValTrue
	return o
}

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Description: `Neon Project.

See details: https://neon.tech/docs/get-started-with-neon/setting-up-a-project/
API: https://api-docs.neon.tech/reference/createproject`,
		SchemaVersion: 9,
		Importer: &schema.ResourceImporter{
			StateContext: resourceProjectImport,
		},
		CreateContext: resourceProjectCreateRetry,
		ReadContext:   resourceProjectReadRetry,
		UpdateContext: resourceProjectUpdateRetry,
		DeleteContext: resourceProjectDeleteRetry,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Project ID.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Identifier of the organisation to which this project belongs.",
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
					supportedVersion := func(v int) bool { return v > 13 && v < 18 }

					if v, ok := i.(int); !ok || !supportedVersion(v) {
						errs = append(
							errs, fmt.Errorf("postgres version %v is not supported", i),
						)
					}

					return
				},
			},
			"store_password": newStoreProjectPasswordDefault(),
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
Note that the feature is available to the Neon Scale plans only. Details: https://neon.tech/docs/manage/projects#configure-ip-allow`,
			},
			"allowed_ips_primary_branch_only": types.NewOptionalTristateBool(
				`Apply the allow-list to the primary branch only.
Note that the feature is available to the Neon Scale plans only.`,
				false),
			"allowed_ips_protected_branches_only": types.NewOptionalTristateBool(
				`Apply the allow-list to the protected branches only.
Note that the feature is available to the Neon Scale plans only.`, false),
			"enable_logical_replication": types.NewOptionalTristateBool(
				`Sets wal_level=logical for all compute endpoints in this project.
All active endpoints will be suspended. Once enabled, logical replication cannot be disabled.
See details: https://neon.tech/docs/introduction/logical-replication
`, true),
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
			"default_endpoint_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default endpoint ID.",
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
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint ID.",
			},
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
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Branch ID.",
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: `The name of the default branch provisioned upon creation of new project. 
If not specified, the default branch name will be used.`,
			},
			"role_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: `The name of the default role provisioned upon creation of new project.
If not specified, the default role name will be used.`,
			},
			"database_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: `The name of the default database provisioned upon creation of new project. It's owned by the default role (` + "`role_name`" + `).
If not specified, the default database name will be used.`,
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
	defaultEndpoint := findDefaultEndpoint(endpoints, branchID)

	defaultDB := findDefaultDatabase(databases, branchID)

	resp, err := c.GetProjectBranchRolePassword(projectID, branchID, defaultDB.OwnerName)
	if err != nil {
		return dbConnectionInfo{}, err
	}
	defaultRolePass := resp.Password

	return dbConnectionInfo{
		userName:   defaultDB.OwnerName,
		pass:       defaultRolePass,
		dbName:     defaultDB.Name,
		host:       defaultEndpoint.Host,
		endpointID: defaultEndpoint.ID,
	}, nil
}

func findDefaultDatabase(databases []neon.Database, defaultBranchID string) neon.Database {
	o := neon.Database{}

	if len(databases) > 0 {
		if len(databases) > 0 {
			var dbs []neon.Database
			for _, el := range databases {
				if el.BranchID == defaultBranchID {
					dbs = append(dbs, el)
				}
			}

			// select the default database based on the creation timestamp
			slices.SortStableFunc(dbs, func(a, b neon.Database) int {
				return a.CreatedAt.Compare(b.CreatedAt)
			})

			o = dbs[0]
		}
	}

	return o
}

func findDefaultEndpoint(endpoints []neon.Endpoint, defaultBranchID string) neon.Endpoint {
	o := neon.Endpoint{}

	if len(endpoints) > 0 {
		var eps []neon.Endpoint
		for _, el := range endpoints {
			// the default endpoint can only be of read_write type
			if !el.Disabled && el.Type == endpointTypeRW && el.BranchID == defaultBranchID {
				eps = append(eps, el)
			}
		}

		// select the default endpoint based on the creation timestamp
		slices.SortStableFunc(eps, func(a, b neon.Endpoint) int {
			return a.CreatedAt.Compare(b.CreatedAt)
		})

		o = eps[0]
	}

	return o
}

type dbConnectionInfo struct {
	userName   string
	pass       string
	dbName     string
	host       string
	endpointID string
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
	if r.OrgID != nil {
		if err := d.Set("org_id", *r.OrgID); err != nil {
			return err
		}
	}
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
	if err := types.SetTristateBool(d, "store_password", &r.StorePasswords); err != nil {
		return err
	}

	defaultEndpointSettings := map[string]interface{}{}
	if v, ok := d.GetOk("default_endpoint_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if v, ok := v.([]interface{})[0].(map[string]interface{}); ok && len(v) > 0 {
			defaultEndpointSettings = v
		}
	}

	if r.DefaultEndpointSettings != nil {
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

	if dbConnectionInfo.endpointID != "" {
		defaultEndpointSettings["id"] = dbConnectionInfo.endpointID
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

		var allowedIPs = make([]string, 0)
		var (
			primaryBranchesOnly   *bool
			protectedBranchesOnly *bool
		)
		if r.Settings.AllowedIps.Ips != nil {
			allowedIPs = *r.Settings.AllowedIps.Ips
			primaryBranchesOnly = r.Settings.AllowedIps.PrimaryBranchOnly
			protectedBranchesOnly = r.Settings.AllowedIps.ProtectedBranchesOnly
		}
		if err := d.Set("allowed_ips", allowedIPs); err != nil {
			return err
		}
		if _, ok := d.GetOk("allowed_ips_primary_branch_only"); ok ||
			(primaryBranchesOnly != nil && *primaryBranchesOnly) {
			if err := types.SetTristateBool(d, "allowed_ips_primary_branch_only", primaryBranchesOnly); err != nil {
				return err
			}
		}
		if _, ok := d.GetOk("allowed_ips_protected_branches_only"); ok ||
			(protectedBranchesOnly != nil && *protectedBranchesOnly) {
			if err := types.SetTristateBool(d, "allowed_ips_protected_branches_only", protectedBranchesOnly); err != nil {
				return err
			}
		}
		if _, ok := d.GetOk("enable_logical_replication"); ok ||
			(r.Settings.EnableLogicalReplication != nil && *r.Settings.EnableLogicalReplication) {
			if err := types.SetTristateBool(d, "enable_logical_replication",
				r.Settings.EnableLogicalReplication); err != nil {
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

	if err := d.Set("default_endpoint_id", dbConnectionInfo.endpointID); err != nil {
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

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "created Project")

	var orgID *string
	if v, ok := d.GetOk("org_id"); ok && v != "" {
		orgID = pointer(v.(string))
	}

	projectDef := neon.ProjectCreateRequestProject{
		OrgID:          orgID,
		Name:           pointer(d.Get("name").(string)),
		Provisioner:    pointer(neon.Provisioner(d.Get("compute_provisioner").(string))),
		RegionID:       pointer(d.Get("region_id").(string)),
		StorePasswords: types.GetTristateBool(d, "store_password"),
	}

	if v, ok := d.Get("history_retention_seconds").(int); ok && v >= 0 {
		projectDef.HistoryRetentionSeconds = pointer(int32(v))
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

		if projectDef.Settings == nil {
			projectDef.Settings = &neon.ProjectSettingsData{}
		}
		projectDef.Settings.AllowedIps = &neon.AllowedIps{
			Ips: &ips,
		}

		projectDef.Settings.AllowedIps.PrimaryBranchOnly = types.GetTristateBool(d,
			"allowed_ips_primary_branch_only")

		projectDef.Settings.AllowedIps.ProtectedBranchesOnly = types.GetTristateBool(d,
			"allowed_ips_protected_branches_only")
	}

	if v := types.GetTristateBool(d, "enable_logical_replication"); v != nil {
		if projectDef.Settings == nil {
			projectDef.Settings = &neon.ProjectSettingsData{}
		}
		projectDef.Settings.EnableLogicalReplication = v
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
		return err
	}

	projectID := resp.ProjectResponse.Project.ID
	d.SetId(projectID)

	branch := resp.BranchResponse.Branch
	info, err := newDbConnectionInfo(client, projectID, branch.ID, resp.EndpointsResponse.Endpoints,
		resp.DatabasesResponse.Databases)
	if err != nil {
		return err
	}

	return updateStateProject(d, resp.ProjectResponse.Project, branch.ID, branch.Name, info)
}

func resourceProjectCreateRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceProjectCreate, ctx, d, meta)
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "update Project")

	req := neon.ProjectUpdateRequest{
		Project: neon.ProjectUpdateRequestProject{},
	}

	if d.HasChange("name") {
		req.Project.Name = pointer(d.Get("name").(string))
	}

	if d.HasChange("history_retention_seconds") {
		req.Project.HistoryRetentionSeconds = pointer(int32(d.Get("history_retention_seconds").(int)))
	}

	if d.HasChange("default_endpoint_settings") {
		if v, ok := d.GetOk("default_endpoint_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			if v, ok := v.([]interface{})[0].(map[string]interface{}); ok && len(v) > 0 {
				req.Project.DefaultEndpointSettings = mapToDefaultEndpointsSettings(v)
			}
		}
	}

	if d.HasChange("quota") {
		if v, ok := d.GetOk("quota"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			if v, ok := v.([]interface{})[0].(map[string]interface{}); ok && len(v) > 0 {
				req.Project.Settings = &neon.ProjectSettingsData{
					Quota: mapToQuotaSettings(v),
				}
			}
		}
	}

	if v, ok := d.GetOk("allowed_ips"); ok && len(v.([]interface{})) > 0 {
		var ips = make([]string, len(v.([]interface{})))
		for i, vv := range v.([]interface{}) {
			ips[i] = fmt.Sprintf("%v", vv)
		}
		if req.Project.Settings == nil {
			req.Project.Settings = &neon.ProjectSettingsData{}
		}
		req.Project.Settings.AllowedIps = &neon.AllowedIps{
			Ips: &ips,
		}
	}

	if req.Project.Settings == nil {
		req.Project.Settings = new(neon.ProjectSettingsData)
	}
	req.Project.Settings.AllowedIps = new(neon.AllowedIps)

	req.Project.Settings.AllowedIps.ProtectedBranchesOnly = types.GetTristateBool(d,
		"allowed_ips_protected_branches_only")

	req.Project.Settings.AllowedIps.PrimaryBranchOnly = types.GetTristateBool(d,
		"allowed_ips_primary_branches_only")

	if req.Project.Settings == nil {
		req.Project.Settings = &neon.ProjectSettingsData{}
	}
	req.Project.Settings.EnableLogicalReplication = types.GetTristateBool(d, "enable_logical_replication")

	_, err := meta.(sdkProject).UpdateProject(d.Id(), req)
	if err != nil {
		return err
	}

	return resourceProjectRead(ctx, d, meta)
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "get Project")
	client := meta.(sdkProject)

	resp, err := client.GetProject(d.Id())
	if err != nil {
		return err
	}

	project := resp.Project

	branches, err := client.ListProjectBranches(d.Id())
	if err != nil {
		return err
	}

	var branchMain neon.Branch
	for _, v := range branches.Branches {
		if v.Primary {
			branchMain = v
			break
		}
	}

	if branchMain.ID == "" {
		return updateStateProject(d, project, branchMain.ID, branchMain.Name, dbConnectionInfo{})
	}

	endpoints, err := client.ListProjectBranchEndpoints(d.Id(), branchMain.ID)
	if err != nil {
		return err
	}

	dbs, err := client.ListProjectBranchDatabases(d.Id(), branchMain.ID)
	if err != nil {
		return err
	}

	info, err := newDbConnectionInfo(client, project.ID, branchMain.ID, endpoints.Endpoints, dbs.Databases)
	if err != nil {
		return err
	}

	return updateStateProject(d, project, branchMain.ID, branchMain.Name, info)
}

func resourceProjectReadRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceProjectRead, ctx, d, meta)
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
	if diags := resourceProjectReadRetry(ctx, d, meta); diags.HasError() {
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
	GrantPermissionToProject(projectID string, cfg neon.GrantPermissionToProjectRequest) (neon.ProjectPermission, error)
	RevokePermissionFromProject(projectID string, permissionID string) (neon.ProjectPermission, error)
	ListProjectPermissions(projectID string) (neon.ProjectPermissions, error)
}
