//go:build !acceptance
// +build !acceptance

package provider

import (
	"context"
	"errors"
	"os"
	"testing"

	neon "github.com/kislerdm/neon-sdk-go"
)

func Test_resourceVPCEndpointAssignmentImport(t *testing.T) {
	if os.Getenv("TF_ACC") == "1" {
		t.Skip("acceptance tests are running")
	}

	t.Parallel()

	const (
		orgID         = "org-test-123"
		regionID      = "aws-eu-central-1"
		vpcEndpointID = "vpce-abcdef123456"
		label         = "my-vpc-endpoint"
		compositeID   = orgID + "/" + regionID + "/" + vpcEndpointID
	)

	t.Run("shall import the vpc endpoint assignment given composite id", func(t *testing.T) {
		resource := resourceVPCEndpointAssignment()
		definition := resource.TestResourceData()
		definition.SetId(compositeID)

		meta := &sdkClientStub{
			stubVPCEndpoint: stubVPCEndpoint{
				VPCEndpointDetails: neon.VPCEndpointDetails{
					Label: label,
				},
			},
		}

		resources, err := resourceVPCEndpointAssignmentImport(context.TODO(), definition, meta)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		d := resources[0]

		if got := d.Get("org_id").(string); got != orgID {
			t.Errorf("org_id = %q, want %q", got, orgID)
		}
		if got := d.Get("region_id").(string); got != regionID {
			t.Errorf("region_id = %q, want %q", got, regionID)
		}
		if got := d.Get("vpc_endpoint_id").(string); got != vpcEndpointID {
			t.Errorf("vpc_endpoint_id = %q, want %q", got, vpcEndpointID)
		}
		if got := d.Get("label").(string); got != label {
			t.Errorf("label = %q, want %q", got, label)
		}

		// Verify ID is set to vpc_endpoint_id only (not composite) for backwards compatibility
		if got := d.Id(); got != vpcEndpointID {
			t.Errorf("id = %q, want %q (vpc_endpoint_id only for backwards compatibility)", got, vpcEndpointID)
		}
	})

	t.Run("shall fail with invalid composite id", func(t *testing.T) {
		resource := resourceVPCEndpointAssignment()
		definition := resource.TestResourceData()
		definition.SetId("invalid-id")

		meta := &sdkClientStub{}

		_, err := resourceVPCEndpointAssignmentImport(context.TODO(), definition, meta)
		if err == nil {
			t.Fatal("error expected")
		}
	})

	t.Run("shall fail when api returns error", func(t *testing.T) {
		resource := resourceVPCEndpointAssignment()
		definition := resource.TestResourceData()
		definition.SetId(compositeID)

		meta := &sdkClientStub{
			stubVPCEndpoint: stubVPCEndpoint{
				err: errors.New("api error"),
			},
		}

		_, err := resourceVPCEndpointAssignmentImport(context.TODO(), definition, meta)
		if err == nil {
			t.Fatal("error expected")
		}
	})
}
