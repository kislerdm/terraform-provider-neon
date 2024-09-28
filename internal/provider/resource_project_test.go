package provider

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
	"github.com/kislerdm/terraform-provider-neon/internal/types"
	"github.com/stretchr/testify/assert"
)

func Test_resourceProjectCreate(t *testing.T) {
	if os.Getenv("TF_ACC") == "1" {
		t.Skip("acceptance tests are running")
	}

	t.Parallel()

	t.Run(
		"shall request project with custom settings", func(t *testing.T) {
			// GIVEN
			meta := &sdkClientStub{}

			resource := resourceProject()

			definition := resource.TestResourceData()
			err := definition.Set("name", "foo")
			if err != nil {
				t.Fatal(err)
			}

			const (
				autoScalingMin        = 0.5
				autoScalingMax        = 2.
				suspendTimeoutSeconds = 100

				quotaActiveTimeSeconds  = 100
				quotaComputeTimeSeconds = 100
				quotaWrittenDataBytes   = 1 << 30
				quotaDataTransferBytes  = quotaWrittenDataBytes * 5
				quotaLogicalSizeBytes   = quotaWrittenDataBytes * 2

				branchName     = "foo"
				branchRoleName = "bar"
				dbName         = "baz"
			)
			var ipsPrimaryBranchOnly = true

			var (
				ips    = []string{"192.168.1.15", "192.168.2.0/20"}
				ipsMap = map[string]struct{}{}
			)

			for _, ip := range ips {
				ipsMap[ip] = struct{}{}
			}

			err = definition.Set(
				"default_endpoint_settings", []interface{}{
					map[string]interface{}{
						"autoscaling_limit_min_cu": autoScalingMin,
						"autoscaling_limit_max_cu": autoScalingMax,
						"suspend_timeout_seconds":  suspendTimeoutSeconds,
					},
				},
			)
			if err != nil {
				t.Fatal(err)
			}

			if err := definition.Set(
				"quota", []interface{}{
					map[string]interface{}{
						"active_time_seconds":  quotaActiveTimeSeconds,
						"compute_time_seconds": quotaComputeTimeSeconds,
						"written_data_bytes":   quotaWrittenDataBytes,
						"data_transfer_bytes":  quotaDataTransferBytes,
						"logical_size_bytes":   quotaLogicalSizeBytes,
					},
				},
			); err != nil {
				t.Fatal(err)
			}

			if err := definition.Set("allowed_ips", ips); err != nil {
				t.Fatal(err)
			}

			if err := types.SetTristateBool(definition, "allowed_ips_primary_branch_only",
				&ipsPrimaryBranchOnly); err != nil {
				t.Fatal(err)
			}

			if err := definition.Set(
				"branch", []interface{}{
					map[string]interface{}{
						"name":          branchName,
						"role_name":     branchRoleName,
						"database_name": dbName,
					},
				},
			); err != nil {
				t.Fatal(err)
			}

			// WHEN
			err = resourceProjectCreate(context.TODO(), definition, meta)

			// THEN
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			v, ok := meta.req.(neon.ProjectCreateRequest)
			if !ok {
				t.Error("unexpected request object type")
			}

			t.Parallel()
			t.Run(
				"shall request custom default_endpoint_settings", func(t *testing.T) {
					settings := v.Project.DefaultEndpointSettings
					if settings == nil {
						t.Fatal("unexpected DefaultEndpointSettings, shall be not nil")
						return
					}

					if *settings.AutoscalingLimitMinCu != autoScalingMin {
						t.Errorf(
							"unexpected AutoscalingLimitMinCu, want: %f, got: %f", autoScalingMin,
							*settings.AutoscalingLimitMinCu,
						)
					}

					if *settings.AutoscalingLimitMaxCu != autoScalingMax {
						t.Errorf(
							"unexpected AutoscalingLimitMaxCu, want: %f, got: %f", autoScalingMax,
							*settings.AutoscalingLimitMaxCu,
						)
					}

					if *settings.SuspendTimeoutSeconds != suspendTimeoutSeconds {
						t.Errorf(
							"unexpected SuspendTimeoutSeconds, want: %d, got: %d", suspendTimeoutSeconds,
							*settings.SuspendTimeoutSeconds,
						)
					}
				},
			)

			t.Run(
				"shall request custom quota", func(t *testing.T) {
					if v.Project.Settings == nil {
						t.Fatal("unexpected Settings, shall be not nil")
						return
					}

					quota := v.Project.Settings.Quota
					if *quota.ActiveTimeSeconds != quotaActiveTimeSeconds {
						t.Errorf(
							"unexpected quota ActiveTimeSeconds, want: %d, got: %d", quotaActiveTimeSeconds,
							quota.ActiveTimeSeconds,
						)
					}

					if *quota.DataTransferBytes != quotaDataTransferBytes {
						t.Errorf(
							"unexpected quota DataTransferBytes, want: %d, got: %d", quotaDataTransferBytes,
							*quota.DataTransferBytes,
						)
					}

					if *quota.LogicalSizeBytes != quotaLogicalSizeBytes {
						t.Errorf(
							"unexpected quota LogicalSizeBytes, want: %d, got: %d", quotaLogicalSizeBytes,
							*quota.LogicalSizeBytes,
						)
					}

					if *quota.WrittenDataBytes != quotaWrittenDataBytes {
						t.Errorf(
							"unexpected quota WrittenDataBytes, want: %d, got: %d", quotaWrittenDataBytes,
							*quota.WrittenDataBytes,
						)
					}

					if *quota.ComputeTimeSeconds != quotaComputeTimeSeconds {
						t.Errorf(
							"unexpected quota ComputeTimeSeconds, want: %d, got: %d", quotaComputeTimeSeconds,
							*quota.ComputeTimeSeconds,
						)
					}
				},
			)

			t.Run(
				"shall request custom default branch", func(t *testing.T) {
					if v.Project.Branch == nil {
						t.Fatal("unexpected Branch, shall be not nil")
						return
					}

					branch := v.Project.Branch
					if !reflect.DeepEqual(branch.Name, pointer(branchName)) {
						t.Errorf("unexpected branch Name, want: %#v, got: %#v", branchName, branch.Name)
					}

					if !reflect.DeepEqual(branch.DatabaseName, pointer(dbName)) {
						t.Errorf(
							"unexpected branch DatabaseName, want: %#v, got: %#v", dbName,
							branch.DatabaseName,
						)
					}

					if !reflect.DeepEqual(branch.RoleName, pointer(branchRoleName)) {
						t.Errorf(
							"unexpected branch RoleName, want: %#v, got: %#v", branchRoleName,
							branch.RoleName,
						)
					}

				},
			)

			t.Run(
				"shall set allowed ips", func(t *testing.T) {
					if v.Project.Settings == nil {
						t.Fatal("unexpected Settings, shall be not nil")
					}

					got := v.Project.Settings.AllowedIps

					var ipsExcess []string
					for _, ip := range *got.Ips {
						if _, ok := ipsMap[ip]; ok {
							delete(ipsMap, ip)
						} else {
							ipsExcess = append(ipsExcess, ip)
						}
					}

					if len(ipsMap) > 0 || len(ipsExcess) > 0 {
						t.Fatalf("unexpected allowed IPs. want = %v, got = %v\n", ips, got)
					}
				},
			)
		},
	)

	t.Run(
		"shall request project with no data retention", func(t *testing.T) {
			// GIVEN
			meta := &sdkClientStub{}

			resource := resourceProject()

			definition := resource.TestResourceData()
			err := definition.Set("name", "foo")
			if err != nil {
				t.Fatal(err)
			}

			err = definition.Set("history_retention_seconds", 0)
			if err != nil {
				t.Fatal(err)
			}

			// WHEN
			err = resourceProjectCreate(context.TODO(), definition, meta)

			// THEN
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			v, ok := meta.req.(neon.ProjectCreateRequest)
			if !ok {
				t.Error("unexpected request object type")
			}

			if v.Project.HistoryRetentionSeconds == nil {
				t.Errorf("HistoryRetentionSeconds must be not nil")
			}

			if v.Project.HistoryRetentionSeconds != nil &&
				*v.Project.HistoryRetentionSeconds != 0 {
				t.Errorf("HistoryRetentionSeconds must be zero")
			}
		},
	)
}

func Test_newDbConnectionInfo(t *testing.T) {
	if os.Getenv("TF_ACC") == "1" {
		t.Skip("acceptance tests are running")
	}

	t.Run("shall use the values with the lowest CreatedAt to define the default database, role and endpoint",
		func(t *testing.T) {
			// see for details: https://github.com/kislerdm/terraform-provider-neon/issues/83
			// GIVEN
			const (
				projectID         = "foo"
				defaultBranchID   = "br-bar"
				defaultEndpointID = "ep-quiet-breeze-a6rnqy6s"
				defaultHost       = defaultEndpointID + ".us-west-2.aws.neon.tech"
				defaultRole       = "r-baz"
				defaultDB         = "db-qux"
				defaultRolePass   = "pass-foo"
			)

			client := &sdkClientStub{
				stubProjectPermission:   stubProjectPermission{},
				stubProjectRolePassword: stubProjectRolePassword{Password: defaultRolePass},
			}

			var (
				endpoints = []neon.Endpoint{
					{
						BranchID:  defaultBranchID,
						ID:        "ep-qux",
						Host:      "ep-qux.us-west-2.aws.neon.tech",
						CreatedAt: time.Now().Add(1 * time.Second),
						UpdatedAt: time.Now().Add(1 * time.Second),
						Disabled:  true,
						Type:      endpointTypeRW,
					},
					{
						BranchID:  defaultBranchID,
						ID:        "ep-baz",
						Host:      "ep-baz.us-west-2.aws.neon.tech",
						CreatedAt: time.Now().Add(2 * time.Second),
						UpdatedAt: time.Now().Add(2 * time.Second),
						Disabled:  false,
						Type:      endpointTypeRW,
					},
					{
						BranchID:  defaultBranchID,
						ID:        defaultEndpointID,
						Host:      defaultHost,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
						Disabled:  false,
						Type:      endpointTypeRW,
					},
				}

				databases = []neon.Database{
					{
						BranchID:  defaultBranchID,
						CreatedAt: time.Now().Add(1 * time.Second),
						UpdatedAt: time.Now().Add(1 * time.Second),
						Name:      "db-foo",
						OwnerName: "r-foo",
					},
					{
						BranchID:  defaultBranchID,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
						Name:      defaultDB,
						OwnerName: defaultRole,
					},
				}
			)

			// WHEN
			got, err := newDbConnectionInfo(client, projectID, defaultBranchID, endpoints, databases)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// THEN
			if got.userName != defaultRole {
				t.Errorf("unexpected userName. want: %s, got: %s", defaultRole, got.userName)
			}

			if got.dbName != defaultDB {
				t.Errorf("unexpected dbName. want: %s, got: %s", defaultDB, got.dbName)
			}

			if got.pass != defaultRolePass {
				t.Errorf("unexpected pass. want: %s, got: %s", defaultRolePass, got.pass)
			}

			if got.host != defaultHost {
				t.Errorf("unexpected host. want: %s, got: %s", defaultHost, got.host)
			}

			if got.endpointID != defaultEndpointID {
				t.Errorf("unexpected endpointID. want: %s, got: %s", defaultEndpointID, got.endpointID)
			}
		})
}

func Test_requestBody_allowed_ips_primary_branch_flag(t *testing.T) {
	tests := map[string]struct {
		projectName                     string
		allowedIPsPrimaryBranchOnly     bool
		wantAllowedIPsPrimaryBranchOnly *bool
	}{
		"shall create project 'bar' with allowed ips, 'allowed_ips_primary_branch_only' is false": {
			projectName:                     "foo",
			allowedIPsPrimaryBranchOnly:     false,
			wantAllowedIPsPrimaryBranchOnly: nil,
		},
		"shall create project 'bar' with allowed ips, 'allowed_ips_primary_branch_only' is true": {
			projectName:                     "foo",
			allowedIPsPrimaryBranchOnly:     true,
			wantAllowedIPsPrimaryBranchOnly: pointer(true),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			meta := &sdkClientStub{}

			resource := resourceProject()
			definition := resource.TestResourceData()
			wantIPs := []string{"192.168.1.15", "192.168.2.0/20"}

			if err := definition.Set("allowed_ips", wantIPs); err != nil {
				t.Fatal(err)
			}

			assert.NoError(t, definition.Set("name", tt.projectName))

			if tt.allowedIPsPrimaryBranchOnly {
				assert.NoError(t,
					types.SetTristateBool(definition,
						"allowed_ips_primary_branch_only",
						&tt.allowedIPsPrimaryBranchOnly),
				)
			}

			assert.NoError(t, resourceProjectCreate(context.TODO(), definition, meta))

			v, ok := meta.req.(neon.ProjectCreateRequest)
			assert.Truef(t, ok, "unexpected request object type")
			assert.Equal(t, tt.projectName, *v.Project.Name)

			got := v.Project.Settings.AllowedIps

			assert.Len(t, *got.Ips, len(wantIPs))
			assert.ElementsMatch(t, wantIPs, *got.Ips)
			assert.Equal(t, tt.wantAllowedIPsPrimaryBranchOnly, got.PrimaryBranchOnly)
		})
	}
}

func Test_resourceProjectCreate_requestBody_allowed_ips_protected_branches_flag(t *testing.T) {
	tests := map[string]struct {
		projectName                         string
		allowedIPsProtectedBranchesOnly     bool
		wantAllowedIPsProtectedBranchesOnly *bool
	}{
		"shall create project 'bar' with allowed ips, 'allowed_ips_protected_branches_only' is false": {
			projectName:                         "foo",
			allowedIPsProtectedBranchesOnly:     false,
			wantAllowedIPsProtectedBranchesOnly: nil,
		},
		"shall create project 'bar' with allowed ips, 'allowed_ips_protected_branches_only' is true": {
			projectName:                         "foo",
			allowedIPsProtectedBranchesOnly:     true,
			wantAllowedIPsProtectedBranchesOnly: pointer(true),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			meta := &sdkClientStub{}
			resource := resourceProject()
			definition := resource.TestResourceData()
			wantIPs := []string{"192.168.1.15", "192.168.2.0/20"}

			if err := definition.Set("allowed_ips", wantIPs); err != nil {
				t.Fatal(err)
			}

			assert.NoError(t, definition.Set("name", tt.projectName))

			if tt.allowedIPsProtectedBranchesOnly {
				assert.NoError(t,
					types.SetTristateBool(definition,
						"allowed_ips_protected_branches_only",
						&tt.allowedIPsProtectedBranchesOnly),
				)
			}

			assert.NoError(t, resourceProjectCreate(context.TODO(), definition, meta))

			v, ok := meta.req.(neon.ProjectCreateRequest)
			assert.Truef(t, ok, "unexpected request object type")
			assert.Equal(t, tt.projectName, *v.Project.Name)

			got := v.Project.Settings.AllowedIps

			assert.Len(t, *got.Ips, len(wantIPs))
			assert.ElementsMatch(t, wantIPs, *got.Ips)
			assert.Equal(t, tt.wantAllowedIPsProtectedBranchesOnly, got.ProtectedBranchesOnly)
		})
	}
}

func Test_resourceProjectUpdate_requestBody_allowed_ips_protected_branches_flag(t *testing.T) {
	wantIPs := []string{"192.168.1.15", "192.168.2.0/20"}

	t.Run("shall set 'allowed_ips_protected_branches_only' to false", func(t *testing.T) {
		meta := &sdkClientStub{}
		resource := resourceProject()
		definition := resource.TestResourceData()

		if err := definition.Set("allowed_ips", wantIPs); err != nil {
			t.Fatal(err)
		}
		assert.NoError(t, definition.Set("name", "Foo"))
		assert.NoError(t, types.SetTristateBool(definition, "allowed_ips_protected_branches_only", pointer(true)))

		assert.NoError(t, resourceProjectCreate(context.TODO(), definition, meta))

		reqCreate, ok := meta.req.(neon.ProjectCreateRequest)
		assert.Truef(t, ok, "unexpected request object type")

		reqCreateIps := reqCreate.Project.Settings.AllowedIps
		assert.ElementsMatch(t, wantIPs, *reqCreateIps.Ips)
		assert.True(t, *reqCreateIps.ProtectedBranchesOnly)

		n := resource.TestResourceData()
		assert.NoError(t, types.SetTristateBool(n, "allowed_ips_protected_branches_only", pointer(false)))

		assert.False(t, *types.GetTristateBool(n, "allowed_ips_protected_branches_only"))

		assert.NoError(t, resourceProjectUpdate(context.TODO(), n, meta))

		reqUpdate, ok := meta.req.(neon.ProjectUpdateRequest)
		assert.Truef(t, ok, "unexpected request object type")

		reqUpdateIps := reqUpdate.Project.Settings.AllowedIps
		assert.False(t, *reqUpdateIps.ProtectedBranchesOnly)
	})
}

func TestValidatePgVersion(t *testing.T) {
	tests := map[string]bool{
		"8":  true,
		"9":  true,
		"10": true,
		"11": true,
		"12": true,
		"13": true,
		"14": false,
		"15": false,
		"16": false,
		"17": false,
		"18": true,
	}

	factories := map[string]func() (*schema.Provider, error){
		"neon": func() (*schema.Provider, error) {
			return newDev(), nil
		},
	}

	t.Parallel()
	for inputVersion, wantErr := range tests {
		t.Run(inputVersion, func(t *testing.T) {
			def := fmt.Sprintf(`resource "neon_project" "this" {pg_version = %s}`, inputVersion)
			var wantErrRegEx *regexp.Regexp
			if wantErr {
				wantErrRegEx = regexp.MustCompile(fmt.Sprintf("postgres version %s is not supported",
					inputVersion))
			}

			testCase := resource.TestCase{
				ProviderFactories: factories,
				Steps: []resource.TestStep{
					{
						Config:             def,
						PlanOnly:           true,
						ExpectNonEmptyPlan: !wantErr,
						ExpectError:        wantErrRegEx,
					},
				},
			}

			resource.UnitTest(t, testCase)
		})
	}

}
