package provider

import (
	"context"
	"net/http"
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
