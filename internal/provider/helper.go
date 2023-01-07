package provider

import (
	"context"
	"errors"
	"time"

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

func isProjectLocked(client neon.Client, projectID string) (bool, error) {
	o, err := client.GetProject(projectID)
	if err != nil {
		return false, err
	}
	return o.Project.Locked, nil
}

type delay struct {
	delay  time.Duration
	maxCnt uint8
}

func (r *delay) Try(client neon.Client, projectID string) bool {
	var i uint8
	for i < r.maxCnt {
		v, err := isProjectLocked(client, projectID)
		if err != nil {
			panic(err)
		}
		if !v {
			return true
		}
		i++
		time.Sleep(r.delay)
	}
	return false
}

func (r *delay) Retry(
	fn func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics,
	ctx context.Context, d *schema.ResourceData, meta interface{},
) diag.Diagnostics {
	var i uint8
	var diags diag.Diagnostics
	for i < r.maxCnt {
		if diags = fn(ctx, d, meta); !diags.HasError() {
			return nil
		}
		i++
		time.Sleep(r.delay)
	}
	return diags
}

var projectReadiness = delay{
	delay:  5 * time.Second,
	maxCnt: 120,
}
