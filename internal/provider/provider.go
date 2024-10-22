package provider

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

type delay struct {
	delay  time.Duration
	maxCnt uint8
}

func (r *delay) Retry(
	fn func(context.Context, *schema.ResourceData, interface{}) error,
	ctx context.Context, d *schema.ResourceData, meta interface{},
) diag.Diagnostics {
	var i uint8
	var err error
	for i < r.maxCnt {
		tflog.Debug(ctx, "API call attempt "+strconv.Itoa(int(i)))

		switch e := fn(ctx, d, meta).(type) {
		case nil:
			return nil
		case neon.Error:
			tflog.Debug(ctx, "API call error code: "+strconv.Itoa(e.HTTPCode))
			switch e.HTTPCode {
			case 200:
				return nil
			case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusLocked:
				tflog.Debug(ctx, "API call delay "+strconv.FormatInt(r.delay.Milliseconds(), 10)+" ms.")
				err = e
				i++
				time.Sleep(r.delay)
			default:
				return diag.FromErr(e)
			}
		default:
			return diag.FromErr(e)
		}
	}
	return diag.FromErr(err)
}

var projectReadiness = delay{
	delay:  1 * time.Second,
	maxCnt: 120,
}

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
	return newWithClient(version, nil)
}

func newWithClient(version string, customHTTPClient neon.HTTPClient) *schema.Provider {
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
		ConfigureContextFunc: configure(version, customHTTPClient),
	}
}

func configure(version string, httpClient neon.HTTPClient) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		cfg := neon.Config{
			Key: d.Get("api_key").(string),
		}

		switch version {
		case "dev":
			cfg.HTTPClient = neon.NewMockHTTPClient()

		default:
			if httpClient == nil {
				// The default timeout is set arbitrary to ensure the connection is shut during resources state update
				httpClient = &http.Client{Timeout: 2 * time.Minute}
			}

			// Wrap the HTTP client into the client which injects the User-Agent header for tracing.
			cfg.HTTPClient = httpClientWithUserAgent{httpClient: httpClient, providerVersion: version}
		}

		c, err := neon.NewClient(cfg)

		return c, diag.FromErr(err)
	}
}

type httpClientWithUserAgent struct {
	httpClient neon.HTTPClient

	providerVersion string
}

const DefaultApplicationName = "kislerdm/neon"

func (h httpClientWithUserAgent) Do(r *http.Request) (*http.Response, error) {
	if r.Header == nil {
		r.Header = make(http.Header)
	}
	r.Header.Set("User-Agent", fmt.Sprintf("tfProvider-%s@%s", DefaultApplicationName, h.providerVersion))
	return h.httpClient.Do(r)
}
