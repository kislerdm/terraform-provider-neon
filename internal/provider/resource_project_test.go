package provider

import (
	"context"
	"reflect"
	"testing"

	neon "github.com/kislerdm/neon-sdk-go"
)

func Test_resourceProjectCreate(t *testing.T) {
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

				ipsPrimaryBranchOnly = true
			)

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

			if err := definition.Set("allowed_ips_primary_branch_only", ipsPrimaryBranchOnly); err != nil {
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
			d := resourceProjectCreate(context.TODO(), definition, meta)

			// THEN
			if d != nil && d.HasError() {
				t.Error("unexpected errors:")
				for _, e := range d {
					t.Error(e.Detail)
				}
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
			d := resourceProjectCreate(context.TODO(), definition, meta)

			// THEN
			if d != nil && d.HasError() {
				t.Error("unexpected errors:")
				for _, e := range d {
					t.Error(e.Detail)
				}
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
