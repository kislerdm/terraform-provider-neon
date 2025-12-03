package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestParseVPCEndpointAssignmentID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    vpcEndpointAssignmentID
		wantErr bool
	}{
		{
			name:  "valid composite ID",
			input: "org-foo-bar-01234567/aws-us-east-1/vpce-1234567890abcdef0",
			want: vpcEndpointAssignmentID{
				OrgID:         "org-foo-bar-01234567",
				RegionID:      "aws-us-east-1",
				VPCEndpointID: "vpce-1234567890abcdef0",
			},
			wantErr: false,
		},
		{
			name:    "missing region",
			input:   "org-foo/vpce-123",
			wantErr: true,
		},
		{
			name:    "just vpc endpoint ID",
			input:   "vpce-1234567890abcdef0",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVPCEndpointAssignmentID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseVPCEndpointAssignmentID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseVPCEndpointAssignmentID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetResourceAttrsFromVPCEndpointAssignmentID(t *testing.T) {
	// Create a ResourceData using the VPC endpoint assignment schema
	resourceSchema := map[string]*schema.Schema{
		"org_id":          {Type: schema.TypeString},
		"region_id":       {Type: schema.TypeString},
		"vpc_endpoint_id": {Type: schema.TypeString},
		"label":           {Type: schema.TypeString},
	}
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{})

	// Parse a composite ID and set attributes
	id := vpcEndpointAssignmentID{
		OrgID:         "org-test-123",
		RegionID:      "aws-eu-central-1",
		VPCEndpointID: "vpce-abcdef123456",
	}
	setResourceAttrsFromVPCEndpointAssignmentID(d, id)

	// Verify attributes were set correctly
	if got := d.Get("org_id").(string); got != id.OrgID {
		t.Errorf("org_id = %q, want %q", got, id.OrgID)
	}
	if got := d.Get("region_id").(string); got != id.RegionID {
		t.Errorf("region_id = %q, want %q", got, id.RegionID)
	}
	if got := d.Get("vpc_endpoint_id").(string); got != id.VPCEndpointID {
		t.Errorf("vpc_endpoint_id = %q, want %q", got, id.VPCEndpointID)
	}
}
