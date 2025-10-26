package provider

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func intValidationNotNegative(v interface{}, s string) (warn []string, errs []error) {
	if vv, ok := v.(int); ok && vv < 0 {
		errs = append(errs, errors.New(s+" must be not negative"))
	}
	return
}

var schemaRegionID = &schema.Schema{
	Type:        schema.TypeString,
	Optional:    true,
	Computed:    true,
	ForceNew:    true,
	Description: "Deployment region: https://neon.tech/docs/introduction/regions",
}

type t interface {
	bool | string | int | int32 | int64 | float64 | float32 | neon.PgVersion | neon.ComputeUnit | neon.Provisioner | neon.EndpointPoolerMode | neon.SuspendTimeoutSeconds
}

func pointer[V t](v V) *V {
	if fmt.Sprintf("%v", v) == "" {
		return nil
	}
	return &v
}

type complexID struct {
	ProjectID, BranchID, Name string
}

func setResourceAttrsFromComplexID(d *schema.ResourceData, r complexID) {
	_ = d.Set("project_id", r.ProjectID)
	_ = d.Set("branch_id", r.BranchID)
	_ = d.Set("name", r.Name)
}

func (v complexID) toString() string {
	return v.ProjectID + "/" + v.BranchID + "/" + v.Name
}

func parseComplexID(s string) (complexID, error) {
	spl := strings.Split(s, "/")
	if len(spl) != 3 {
		return complexID{}, errors.New(
			"ID of this resource type shall follow the template: {{.ProjectID}}/{{.BranchID}}/{{.Name}}",
		)
	}
	return complexID{
		ProjectID: spl[0],
		BranchID:  spl[1],
		Name:      spl[2],
	}, nil
}
