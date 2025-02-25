package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	neon "github.com/kislerdm/neon-sdk-go"
)

func TestAccOrg(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1")
	}

	orgID := os.Getenv("ORG_ID")
	if orgID == "" {
		t.Skip("ORG_ID must be set")
	}

	client, err := neon.NewClient(neon.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		resp, _ := client.ListProjects(nil, nil, &projectNamePrefix, &orgID, nil)
		for _, project := range resp.Projects {
			_, _ = client.DeleteProject(project.ID)
		}
	})

	projectName := newProjectName()

	resourceDefinition := fmt.Sprintf(`resource "neon_project" "this" {
    org_id              = "%s"
	name                = "%s"
}`, orgID, projectName)

	resource.Test(
		t, resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"neon": func() (*schema.Provider, error) {
					return newAccTest(), nil
				},
			},
			Steps: []resource.TestStep{
				{
					Config: resourceDefinition,
					Check: func(state *terraform.State) error {
						var (
							e    error
							resp neon.ListProjectsRespObj
						)
						resp, e = client.ListProjects(nil, nil, &projectName, &orgID, nil)
						if e == nil {
							if len(resp.Projects) != 1 {
								e = fmt.Errorf(
									"project %s should have been creted in the org %s", projectName, orgID,
								)
							}
						}
						return e
					},
				},
			},
		},
	)
}

func TestAccVPCEndpoint(t *testing.T) {
	t.Skip("cannot be automated yet")

	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1")
	}

	orgID := os.Getenv("ORG_ID")
	if orgID == "" {
		t.Skip("ORG_ID must be set")
	}

	neonClient, err := neon.NewClient(neon.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		t.Fatal(err)
	}

	const neonServiceName = "com.amazonaws.vpce.eu-central-1.vpce-svc-05554c35009a5eccb"
	const region = "eu-central-1"

	/* FIXME(?) the following error is returned on attempt to provisions the vpc endpoint
	    following the Neon instructions: https://neon.tech/docs/guides/neon-private-networking

	Error: creating EC2 VPC Endpoint (com.amazonaws.vpce.eu-central-1.vpce-svc-05554c35009a5eccb):
	operation error EC2: CreateVpcEndpoint, https response error StatusCode: 400,
	RequestID: f3b0c1e6-ba70-4806-a596-a5c4ae8352ce,
	api error
	InvalidServiceName: The Vpc Endpoint Service 'com.amazonaws.vpce.eu-central-1.vpce-svc-05554c35009a5eccb' does not exist
	*/

	awsState := fmt.Sprintf(`resource "aws_vpc_endpoint" "_" {
		vpc_id              = "vpc-1697377c"
		service_name        = "%s"
		vpc_endpoint_type   = "Interface"
		subnet_ids          = ["subnet-02a2774e"]
		ip_address_type     = "ipv4"
		private_dns_enabled = false
	}`, neonServiceName)

	vpcEndpointID := "vpce-0859f03d7c4c35f4c"

	vpcAssignment := fmt.Sprintf(`%s
resource "neon_vpc_endpoint_assignment" "_" {
    org_id          = "%s"
	region_id       = "aws-%s"
	vpc_endpoint_id = "%s"
	label           = "foo"
}`, awsState, orgID, region, vpcEndpointID)

	projectName := newProjectName()
	vpcRestriction := fmt.Sprintf(`resource "neon_project" "_" {
	name      = "%s"
	region_id = "aws-%s"
	org_id    = "%s"
}

resource "neon_vpc_endpoint_restriction" "_" {
	project_id      = neon_project._.id
	vpc_endpoint_id = "%s"
	label           = "bar"
}`, projectName, region, orgID, vpcEndpointID)

	var projectID string

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"neon": func() (*schema.Provider, error) {
					return newAccTest(), nil
				},
			},
			ExternalProviders: map[string]resource.ExternalProvider{
				"aws": {
					VersionConstraint: "5.88.0",
					Source:            "hashicorp/aws",
				},
			},
			Steps: []resource.TestStep{
				{
					Config:  vpcAssignment,
					Destroy: false,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("neon_vpc_endpoint_assignment._",
							"vpc_endpoint_id", vpcEndpointID),
						resource.TestCheckResourceAttr("neon_vpc_endpoint_assignment._",
							"id", vpcEndpointID),
					),
				},
				{
					Config:        vpcAssignment,
					ImportState:   true,
					ResourceName:  "neon_vpc_endpoint_assignment._",
					ImportStateId: vpcEndpointID,
					Destroy:       false,
				},
				{
					Config:  vpcRestriction,
					Destroy: false,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("neon_vpc_endpoint_restriction._",
							"vpc_endpoint_id", vpcEndpointID),
						func(s *terraform.State) error {
							pr, err := neonClient.ListProjects(nil, nil, &projectName, &orgID, nil)
							if err == nil {
								for _, project := range pr.Projects {
									if projectName == project.Name {
										projectID = project.ID
										break
									}
								}
							}
							if projectID == "" {
								err = fmt.Errorf("no project %s found", projectName)
							}
							if err == nil {
								wantID := fmt.Sprintf("%s/%s", vpcEndpointID, projectID)
								err = resource.TestCheckResourceAttr(
									"neon_vpc_endpoint_restriction._", "id", wantID,
								)(s)
							}
							return err
						},
					),
				},
				{
					Config: fmt.Sprintf(`resource "neon_vpc_endpoint_restriction" "_" {
	project_id      = "%s"
	vpc_endpoint_id = "%s"
	label           = "bar"
}`, projectID, vpcEndpointID),
					ImportState:   true,
					ResourceName:  "neon_vpc_endpoint_restriction._",
					ImportStateId: fmt.Sprintf("%s/%s", vpcEndpointID, projectID),
				},
			},
		},
	)
}
