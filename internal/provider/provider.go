package provider

import (
	"context"
	"math/rand"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func init() {
	rand.New(rand.NewSource(time.Now().Unix()))
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown
}

func newDev() *schema.Provider {
	return New("dev")
}

func New(version string) *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "API access key. Default is read from the environment variable `NEON_API_KEY`.",
				Default:     os.Getenv("NEON_API_KEY"),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"neon_project":            resourceProject(),
			"neon_branch":             resourceBranch(),
			"neon_endpoint":           resourceEndpoint(),
			"neon_role":               resourceRole(),
			"neon_database":           resourceDatabase(),
			"neon_project_permission": resourceProjectPermission(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"neon_project":              dataSourceProject(),
			"neon_branches":             dataSourceBranches(),
			"neon_branch_endpoints":     dataSourceBranchEndpoints(),
			"neon_branch_roles":         dataSourceBranchRoles(),
			"neon_branch_role_password": dataSourceBranchRolePassword(),
		},
		ConfigureContextFunc: configure(version),
	}
}

func configure(version string) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		if version == "dev" {
			c, err := neon.NewClient(neon.Config{HTTPClient: neon.NewMockHTTPClient()})
			if err != nil {
				return nil, diag.FromErr(err)
			}
			return c, diag.FromErr(err)
		}
		c, err := neon.NewClient(neon.Config{Key: d.Get("api_key").(string)})
		if err != nil {
			return nil, diag.FromErr(err)
		}
		return c, nil
	}
}
