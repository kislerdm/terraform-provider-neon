package provider

import (
	"context"
	"math/rand"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
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
		"neon_api_key":                  resourceAPIKey(),
		"neon_project":                  resourceProject(),
		"neon_branch":                   resourceBranch(),
		"neon_endpoint":                 resourceEndpoint(),
		"neon_role":                     resourceRole(),
		"neon_database":                 resourceDatabase(),
		"neon_project_permission":       resourceProjectPermission(),
		"neon_jwks_url":                 resourceJwksUrl(),
		"neon_vpc_endpoint_assignment":  resourceVPCEndpointAssignment(),
		"neon_vpc_endpoint_restriction": resourceVPCEndpointRestriction(),
		"neon_org_api_key":              resourceOrgAPIKey(),
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

func newAccTest() *schema.Provider {
	return New("accTest")
}

// NewFramework returns the Framework provider used by the protocol mux.
func NewFramework(version string) frameworkprovider.Provider {
	return &frameworkProvider{version: version}
}

var _ frameworkprovider.Provider = (*frameworkProvider)(nil)

// frameworkProvider is the Framework portion of the provider. It is served
// through terraform-plugin-mux alongside the legacy SDK provider.
type frameworkProvider struct {
	version string
}

type frameworkProviderConfigModel struct {
	APIKey types.String `tfsdk:"api_key"`
}

func (p *frameworkProvider) Metadata(_ context.Context, _ frameworkprovider.MetadataRequest, resp *frameworkprovider.MetadataResponse) {
	resp.TypeName = "neon"
	resp.Version = p.version
}

func (p *frameworkProvider) Schema(_ context.Context, _ frameworkprovider.SchemaRequest, resp *frameworkprovider.SchemaResponse) {
	resp.Schema = providerschema.Schema{
		Attributes: map[string]providerschema.Attribute{
			"api_key": providerschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "API access key. Default is read from the environment variable `NEON_API_KEY`.",
			},
		},
	}
}

func (p *frameworkProvider) Configure(ctx context.Context, req frameworkprovider.ConfigureRequest,
	resp *frameworkprovider.ConfigureResponse) {
	var config frameworkProviderConfigModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := config.APIKey.ValueString()
	if key == "" {
		key = os.Getenv("NEON_API_KEY")
	}

	cfg := neon.Config{
		Key:        key,
		HTTPClient: telemetry.NewHTTPClient(Name, p.version, req.TerraformVersion),
	}
	client, err := neon.NewClient(cfg)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Configure Neon Provider", err.Error())
		return
	}

	resp.ResourceData = client
}

func (p *frameworkProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewBranchBackupScheduleResource,
	}
}

func (p *frameworkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

func NewServer(version string) (tfprotov6.ProviderServer, error) {
	legacyServer, err := tf5to6server.UpgradeServer(context.Background(), func() tfprotov5.ProviderServer {
		return New(version).GRPCProvider()
	})
	if err != nil {
		return nil, err
	}
	return tf6muxserver.NewMuxServer(context.Background(),
		func() tfprotov6.ProviderServer {
			return legacyServer
		},
		providerserver.NewProtocol6(NewFramework(version)),
	)
}

func newAccTestFramework() tfprotov6.ProviderServer {
	o, err := NewServer("accTest")
	if err != nil {
		panic(err)
	}
	return o
}
