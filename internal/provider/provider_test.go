//go:build !acceptance
// +build !acceptance

package provider

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

type httpClient struct {
	header http.Header
}

func (h *httpClient) Do(r *http.Request) (*http.Response, error) {
	h.header = r.Header
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("{}")),
	}, nil
}

func TestUserAgentInstrumentation(t *testing.T) {
	const (
		wantVersion = "ProviderVer"

		resourceDefinition = `resource "neon_role" "this" {
	project_id = "foo"
	branch_id  = "foo"
	name       = "foo"
}`
	)

	c := &httpClient{}

	t.Setenv("NEON_API_KEY", "foo")

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"neon": func() (*schema.Provider, error) {
				return newWithClient(wantVersion, c), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config:             resourceDefinition,
				ExpectNonEmptyPlan: true,
			},
		},
	})

	userAgent := c.header.Get("User-Agent")
	assert.Contains(t, userAgent, DefaultApplicationName)
	els := strings.Split(userAgent, "@")
	assert.Len(t, els, 2)
	assert.Equal(t, wantVersion, els[1])
}
