package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

var (
	endpointTypeRW       = neon.EndpointTypeReadWrite
	endpointTypeReadOnly = neon.EndpointTypeReadOnly
)

func resourceEndpoint() *schema.Resource {
	return &schema.Resource{
		Description:   `Project Endpoint. See details: https://neon.tech/docs/manage/endpoints/`,
		SchemaVersion: 8,
		Importer: &schema.ResourceImporter{
			StateContext: resourceEndpointImport,
		},
		CreateContext: resourceEndpointCreateRetry,
		ReadContext:   resourceEndpointReadRetry,
		UpdateContext: resourceEndpointUpdateRetry,
		DeleteContext: resourceEndpointDeleteRetry,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint ID.",
			},
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Project ID.",
			},
			"branch_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Branch ID.",
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     endpointTypeRW.String(),
				Description: `Access type. **Note** that a single branch can have only one "read_write" endpoint.`,
				ValidateFunc: func(d interface{}, k string) (warn []string, errs []error) {
					switch v := d.(string); v {
					case endpointTypeRW.String(), endpointTypeReadOnly.String():
					default:
						errs = append(errs, errors.New(v+" is not supported value for "+k))
					}
					return warn, errs
				},
			},
			"host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint URI.",
			},
			"region_id": schemaRegionID,
			"autoscaling_limit_min_cu": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Computed:    true,
				Description: "Minimal value of the compute autoscaling limit.",
			},
			"autoscaling_limit_max_cu": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Computed:    true,
				Description: "Maximal value of the compute autoscaling limit.",
			},
			"pg_settings": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Disable the endpoint.",
			},
			"proxy_host": {
				Type:     schema.TypeString,
				Computed: true,
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
			"suspend_timeout_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				Description: `Duration of inactivity in seconds after which the compute endpoint is automatically suspended.
The value 0 means use the global default.
The value -1 means never suspend. The default value is 300 seconds (5 minutes).
The maximum value is 604800 seconds (1 week)`,
				ValidateFunc: func(d interface{}, k string) (_ []string, errs []error) {
					var v int64
					switch d := d.(type) {
					case int:
						v = int64(d)
					case int64:
						v = d
					}
					if v > 604800 || v < -1 {
						errs = append(errs, fmt.Errorf("%d is not supported value for %s", v, k))
					}
					return nil, errs
				},
			},
			"host_pooling": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint URI for connection pooling.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Compute name.",
			},
		},
	}
}

func updateStateEndpoint(d *schema.ResourceData, v neon.Endpoint) error {
	if err := d.Set("type", v.Type.String()); err != nil {
		return err
	}
	if err := d.Set("host", v.Host); err != nil {
		return err
	}
	if err := d.Set("host_pooling", newPooledHost(v.Host)); err != nil {
		return err
	}
	if err := d.Set("region_id", v.RegionID); err != nil {
		return err
	}
	if err := d.Set("autoscaling_limit_min_cu", float64(v.AutoscalingLimitMinCu)); err != nil {
		return err
	}
	if err := d.Set("autoscaling_limit_max_cu", float64(v.AutoscalingLimitMaxCu)); err != nil {
		return err
	}
	if v.Settings.PgSettings != nil {
		if err := d.Set("pg_settings", v.Settings.PgSettings); err != nil {
			return err
		}
	}
	if err := d.Set("disabled", v.Disabled); err != nil {
		return err
	}
	if err := d.Set("proxy_host", v.ProxyHost); err != nil {
		return err
	}
	if err := d.Set("compute_provisioner", string(v.Provisioner)); err != nil {
		return err
	}
	if err := d.Set("suspend_timeout_seconds", int64(v.SuspendTimeoutSeconds)); err != nil {
		return err
	}
	if err := d.Set("branch_id", v.BranchID); err != nil {
		return err
	}
	if v.Name != nil {
		if err := d.Set("name", *v.Name); err != nil {
			return err
		}
	}
	return nil
}

func resourceEndpointCreateRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceEndpointCreate, ctx, d, meta)
}

func resourceEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "created Endpoint")

	endpointType, err := neon.NewEndpointType(d.Get("type").(string))
	if err != nil {
		return err
	}

	cfg := neon.EndpointCreateRequestEndpoint{
		BranchID:    d.Get("branch_id").(string),
		Type:        endpointType,
		RegionID:    pointer(d.Get("region_id").(string)),
		Disabled:    pointer(d.Get("disabled").(bool)),
		Provisioner: pointer(neon.Provisioner(d.Get("compute_provisioner").(string))),
	}

	if v, ok := d.GetOk("autoscaling_limit_min_cu"); ok {
		cfg.AutoscalingLimitMinCu = pointer(neon.ComputeUnit(v.(float64)))
	}

	if v, ok := d.GetOk("autoscaling_limit_max_cu"); ok {
		cfg.AutoscalingLimitMaxCu = pointer(neon.ComputeUnit(v.(float64)))
	}

	if v, ok := d.GetOk("suspend_timeout_seconds"); ok {
		cfg.SuspendTimeoutSeconds = pointer(neon.SuspendTimeoutSeconds(v.(int)))
	}

	if v, ok := d.GetOk("pg_settings"); ok && len(v.(map[string]any)) > 0 {
		var pgSettings = make(neon.PgSettingsData, len(v.(map[string]any)))
		for k, vv := range v.(map[string]any) {
			pgSettings[k] = vv
		}
		cfg.Settings = &neon.EndpointSettingsData{
			PgSettings: &pgSettings,
		}
	}

	if v, ok := d.GetOk("name"); ok && v.(string) != "" {
		cfg.Name = pointer(v.(string))
	}

	client := meta.(*neon.Client)
	resp, err := client.CreateProjectEndpoint(
		d.Get("project_id").(string),
		neon.EndpointCreateRequest{Endpoint: cfg},
	)
	if err != nil {
		return err
	}

	waitUnfinishedOperations(ctx, client, resp.OperationsResponse.Operations)

	d.SetId(resp.Endpoint.ID)

	return updateStateEndpoint(d, resp.EndpointResponse.Endpoint)
}

func resourceEndpointReadRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.RetryWithFallback(resourceEndpointRead, ctx, d, meta,
		map[int]FallbackFn{
			http.StatusNotFound: func(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
				tflog.Debug(ctx, "endpoint not found, removing from state",
					map[string]interface{}{
						"id":         d.Id(),
						"project_id": d.Get("project_id"),
						"branch_id":  d.Get("branch_id"),
					})
				d.SetId("")
				return nil
			}})
}

func resourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "read Endpoint")

	resp, err := meta.(*neon.Client).GetProjectEndpoint(
		d.Get("project_id").(string),
		d.Id(),
	)
	if err != nil {
		return err
	}

	return updateStateEndpoint(d, resp.Endpoint)
}

func resourceEndpointUpdateRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceEndpointUpdate, ctx, d, meta)
}

func resourceEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "update Endpoint")

	cfg := neon.EndpointUpdateRequestEndpoint{
		Disabled:              pointer(d.Get("disabled").(bool)),
		BranchID:              pointer(d.Get("branch_id").(string)),
		AutoscalingLimitMinCu: pointer(neon.ComputeUnit(d.Get("autoscaling_limit_min_cu").(float64))),
		AutoscalingLimitMaxCu: pointer(neon.ComputeUnit(d.Get("autoscaling_limit_max_cu").(float64))),
		Provisioner:           pointer(neon.Provisioner(d.Get("compute_provisioner").(string))),
	}

	if d.HasChange("suspend_timeout_seconds") {
		if v, ok := d.GetOk("suspend_timeout_seconds"); ok {
			cfg.SuspendTimeoutSeconds = pointer(neon.SuspendTimeoutSeconds(v.(int)))
		}
	}

	if d.HasChange("pg_settings") {
		if v, ok := d.GetOk("pg_settings"); ok && len(v.(map[string]any)) > 0 {
			var pgSettings = make(neon.PgSettingsData, len(v.(map[string]any)))
			for k, vv := range v.(map[string]any) {
				pgSettings[k] = vv
			}
			cfg.Settings = &neon.EndpointSettingsData{
				PgSettings: &pgSettings,
			}
		}
	}

	if d.HasChange("name") && d.Get("name").(string) != "" {
		cfg.Name = pointer(d.Get("name").(string))
	}

	client := meta.(*neon.Client)
	resp, err := client.UpdateProjectEndpoint(
		d.Get("project_id").(string),
		d.Id(),
		neon.EndpointUpdateRequest{Endpoint: cfg},
	)
	if err != nil {
		return err
	}
	waitUnfinishedOperations(ctx, client, resp.OperationsResponse.Operations)
	return updateStateEndpoint(d, resp.EndpointResponse.Endpoint)
}

func resourceEndpointImport(ctx context.Context, d *schema.ResourceData, meta interface{}) (
	[]*schema.ResourceData, error,
) {
	tflog.Trace(ctx, "import Endpoint")
	tflog.Debug(ctx, "split input ID")
	els := strings.SplitN(d.Id(), "/", 2)
	if len(els) != 2 {
		return nil, fmt.Errorf("invalid format of provided identifier, expected: {{.ProjectID}}/{{.EndpointID}}")
	}

	projectID := els[0]
	d.SetId(els[1])
	if err := d.Set("project_id", projectID); err != nil {
		return nil, err
	}
	if diags := projectReadiness.Retry(resourceEndpointRead, ctx, d, meta); diags.HasError() {
		d.SetId("")
		return nil, errors.New(diags[0].Summary)
	}
	return []*schema.ResourceData{d}, nil
}

func resourceEndpointDeleteRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.RetryWithFallback(resourceEndpointDelete, ctx, d, meta, map[int]FallbackFn{
		http.StatusNotFound: func(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
			d.SetId("")
			return nil
		},
		http.StatusUnprocessableEntity: func(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
			d.SetId("")
			return nil
		},
	})
}

func resourceEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "delete Endpoint")
	client := meta.(*neon.Client)
	op, err := client.DeleteProjectEndpoint(d.Get("project_id").(string), d.Id())
	if err != nil {
		return err
	}
	waitUnfinishedOperations(ctx, client, op.OperationsResponse.Operations)
	d.SetId("")
	return updateStateEndpoint(d, neon.Endpoint{})
}
