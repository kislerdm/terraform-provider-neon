package provider

import (
	"errors"
	"fmt"

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

func mapToPgSettings(v map[string]interface{}) *neon.PgSettingsData {
	if len(v) == 0 {
		return nil
	}
	o := make(neon.PgSettingsData, len(v))
	for k, v := range v {
		o[k] = v
	}
	return &o
}

func intValidationNotNegative(v interface{}, s string) (warn []string, errs []error) {
	if v.(int) < 0 {
		errs = append(errs, errors.New(s+" must be not negative"))
		return
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

func validateAutoscallingLimit(val interface{}, name string) (warns []string, errs []error) {
	switch val.(type) {
	case float64:
		switch v := val.(float64); v {
		case 0.25,
			0.5,
			1.,
			2.,
			3.,
			4.,
			5.,
			6.,
			7.:
			return
		}
	case int:
		switch v := val.(int); v {
		case 1, 2, 3, 4, 5, 6, 7:
			return
		}
	}
	errs = append(
		errs, fmt.Errorf(
			`%v is unsopported value for %s, 
details: https://neon.tech/docs/manage/endpoints#compute-size-and-autoscaling-configuration`, val, name,
		),
	)
	return
}
