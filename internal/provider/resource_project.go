package provider

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/kislerdm/neon-sdk-go"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Description: "Neon Project. See details: https://neon.tech/docs/get-started-with-neon/setting-up-a-project/",

		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		UpdateContext: resourceProjectUpdate,
		DeleteContext: resourceProjectDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     uuid.NewString(),
				Description: "Project name.",
			},
			"platform_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "serverless",
				Description: "Platform type id.",
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					if v, _ := i.(string); v != "serverless" {
						return diag.Errorf("platform_id is not recognised.")
					}
					return nil
				},
			},
			"region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "us-west-2",
				Description: "AWS Region.",
			},
			"instance_handle": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "scalable",
				Description: "Instance type name.",
			},
			"settings": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Project custom settings.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Project ID.",
			},
			"instance_type_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Instance type ID.",
			},
			"platform_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Platform type name.",
			},
			"region_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "AWS Region name.",
			},
			"parent_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Project parent.",
			},
			"roles": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of roles for the project.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Role ID.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Role name.",
						},
						"password": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Role access password.",
							Sensitive:   true,
						},
					},
				},
			},
			"databases": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of the project databases.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Role ID.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Role name.",
						},
						"owner_id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Owner role ID.",
						},
					},
				},
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Project creation timestamp.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Project last update timestamp.",
			},
			"pending_state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Project pending state.",
			},
			"current_state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Project current state.",
			},
			"deleted": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flag is the project is deleted.",
			},
			"size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Project size.",
			},
			"max_project_size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Project max size.",
			},
			"pooler_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flag if pooler is enabled.",
			},
		},
	}
}

func updateProjectInfoState(d *schema.ResourceData, resp sdk.ProjectInfo) {
	_ = d.Set("name", resp.Name)
	_ = d.Set("platform_id", resp.PlatformID)
	_ = d.Set("region_id", resp.RegionID)
	_ = d.Set("instance_handle", resp.InstanceHandle)
	_ = d.Set("settings", resp.Settings)
	_ = d.Set("id", resp.ID)
	_ = d.Set("instance_type_id", resp.InstanceTypeID)
	_ = d.Set("platform_name", resp.PlatformName)
	_ = d.Set("region_name", resp.RegionName)
	_ = d.Set("parent_id", resp.ParentID)
	_ = d.Set("roles", resp.Roles)
	_ = d.Set("databases", resp.Databases)
	_ = d.Set("created_at", resp.CreatedAt)
	_ = d.Set("updated_at", resp.UpdatedAt)
	_ = d.Set("pending_state", resp.PendingState)
	_ = d.Set("current_state", resp.CurrentState)
	_ = d.Set("deleted", resp.Deleted)
	_ = d.Set("size", resp.Size)
	_ = d.Set("max_project_size", resp.MaxProjectSize)
	_ = d.Set("pooler_enabled", resp.PoolerEnabled)
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Trace(ctx, "created Project")
	client := meta.(sdk.Client)

	resp, err := client.CreateProject(
		sdk.ProjectSettingsRequestCreate{
			Name:           d.Get("name").(string),
			PlatformID:     d.Get("platform_id").(string),
			RegionID:       d.Get("region_id").(string),
			InstanceHandle: d.Get("instance_handle").(string),
			Settings:       d.Get("settings").(map[string]string),
		},
	)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d", rand.Int()))
	updateProjectInfoState(d, resp)

	return nil
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Trace(ctx, "update Project")
	resp, err := meta.(sdk.Client).UpdateProject(
		sdk.ProjectSettingsRequestUpdate{
			InstanceTypeID: d.Get("instance_type_id").(string),
			Name:           d.Get("name").(string),
			PoolerEnabled:  d.Get("pooler_enabled").(bool),
			Settings:       d.Get("settings").(map[string]string),
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	updateProjectInfoState(d, resp)

	return nil
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(sdk.Client)
	resp, err := client.ReadInfoProject(d.Get("id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	updateProjectInfoState(d, resp)

	return nil
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(sdk.Client)
	if _, err := client.DeleteProject(d.Get("id").(string)); err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")

	updateProjectInfoState(d, sdk.ProjectInfo{})

	return nil
}
