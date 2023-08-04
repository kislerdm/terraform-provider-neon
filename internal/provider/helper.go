package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func pgSettingsToMap(v neon.PgSettingsData) map[string]interface{} {
	o := make(map[string]interface{}, len(v))
	for k, v := range v {
		o[k] = v
	}
	return o
}

func mapToPgSettings(v map[string]interface{}) neon.PgSettingsData {
	if len(v) == 0 {
		return nil
	}
	o := make(neon.PgSettingsData, len(v))
	for k, v := range v {
		o[k] = v
	}
	return o
}

func intValidationNotNegative(v interface{}, s string) (warn []string, errs []error) {
	if v.(int) < 0 {
		errs = append(errs, errors.New(s+" must be not negative"))
		return
	}
	return
}

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
	delay:  5 * time.Second,
	maxCnt: 120,
}

var schemaRegionID = &schema.Schema{
	Type:        schema.TypeString,
	Optional:    true,
	Computed:    true,
	ForceNew:    true,
	Description: "AWS Region.",
	ValidateFunc: func(i interface{}, s string) (warns []string, errs []error) {
		switch v := i.(string); v {
		case "aws-us-east-2", "aws-us-west-2", "aws-eu-central-1", "aws-ap-southeast-1":
			return
		default:
			errs = append(
				errs,
				errors.New(
					"region "+v+" is not supported yet: https://neon.tech/docs/introduction/regions/",
				),
			)
			return
		}
	},
}

type complexID struct {
	ProjectID, BranchID, Name string
}

func setResourceDataFromComplexID(d *schema.ResourceData, r complexID) {
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

type t interface {
	bool | string | int | int32 | int64 | float64 | float32 | neon.PgVersion | neon.ComputeUnit | neon.Provisioner | neon.EndpointPoolerMode
}

func pointer[V t](v V) *V {
	if fmt.Sprintf("%v", v) == "" {
		return nil
	}
	return &v
}

func validateAutoscallingLimit(val interface{}, name string) (warns []string, errs []error) {
	switch v := val.(float64); v {
	case
		0.25,
		0.5,
		1.,
		2.,
		3.,
		4.,
		5.,
		6.,
		7.:
		return
	default:
		errs = append(
			errs, fmt.Errorf(
				`%v is unsopported value for %s, 
details: https://neon.tech/docs/manage/endpoints#compute-size-and-autoscaling-configuration`, v, name,
			),
		)
		return
	}
}
