package provider

import (
	"context"
	"math/rand"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
	"github.com/kislerdm/terraform-provider-neon/provider/telemetry"
)

const Name = "kislerdm/neon"

func init() {
	rand.New(rand.NewSource(time.Now().Unix()))
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown
}

var p = &schema.Provider{
	Schema: map[string]*schema.Schema{
		"api_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "API access key. Default is read from the environment variable `NEON_API_KEY`.",
			Default:     os.Getenv("NEON_API_KEY"),
		},
	},
	ResourcesMap: map[string]*schema.Resource{
		"neon_api_key":                 resourceAPIKey(),
		"neon_project":                 resourceProject(),
		"neon_branch":                  resourceBranch(),
		"neon_endpoint":                resourceEndpoint(),
		"neon_role":                    resourceRole(),
		"neon_database":                resourceDatabase(),
		"neon_project_permission":      resourceProjectPermission(),
		"neon_jwks_url":                resourceJwksUrl(),
		"neon_vpc_endpoint_assignment": resourceVPCEndpointAssignment(),
	},
	DataSourcesMap: map[string]*schema.Resource{
		"neon_project":              dataSourceProject(),
		"neon_branches":             dataSourceBranches(),
		"neon_branch_endpoints":     dataSourceBranchEndpoints(),
		"neon_branch_roles":         dataSourceBranchRoles(),
		"neon_branch_role_password": dataSourceBranchRolePassword(),
	},
}

// New returns the provider.
func New(version string) *schema.Provider {
	var o = new(schema.Provider)
	*o = *p
	o.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (c interface{},
		errs diag.Diagnostics) {
		var err error
		c, err = neon.NewClient(neon.Config{
			Key:        d.Get("api_key").(string),
			HTTPClient: telemetry.NewHTTPClient(Name, version, o.TerraformVersion),
		})
		if err != nil {
			errs = diag.FromErr(err)
		}
		return c, errs
	}
	return o
}

// NewUnitTest returns the provider's factory for unit tests.
func NewUnitTest() *schema.Provider {
	var o = new(schema.Provider)
	*o = *p
	o.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return neon.NewMockHTTPClient(), nil
	}
	return o
}

func newAccTest() *schema.Provider {
	return New("accTest")
}
