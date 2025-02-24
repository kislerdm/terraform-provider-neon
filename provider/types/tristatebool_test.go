//go:build !acceptance
// +build !acceptance

package types

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestNewOptionalTristateBool(t *testing.T) {
	type args struct {
		description string
		forceNew    bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "shall create attr which does not force recreation and contains only header in description",
			args: args{
				description: "",
				forceNew:    false,
			},
		},
		{
			name: "shall create attr which forces recreation and contains only header in description",
			args: args{
				description: "",
				forceNew:    true,
			},
		},
		{
			name: "shall create attr which forces recreation and custom description",
			args: args{
				description: "foo",
				forceNew:    true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOptionalTristateBool(tt.args.description, tt.args.forceNew)

			assert.True(t, got.Optional, "shall set Option to true")

			wantForceNew := tt.args.forceNew
			assert.Equal(t, wantForceNew, got.ForceNew, fmt.Sprintf("shell set ForceNew to %v", wantForceNew))

			assert.Equal(t, schema.TypeString, got.Type, "shall be of the TypeString")

			if !strings.HasPrefix(got.Description,
				"Set to 'yes' to activate, 'no' to deactivate explicitly, and omit to keep the default value.\n") {
				t.Error("unexpected Description's header")
			}
			if !strings.HasSuffix(got.Description, tt.args.description) {
				t.Error("unexpected Description's ending")
			}
		})
	}
}

func Test_validateFuncNewOptionalTristateBool(t *testing.T) {
	const key = "attr_name"

	tests := []struct {
		name     string
		val      interface{}
		wantErrs []error
	}{
		{
			name: "shall return no errors given 'yes' as input",
			val:  "yes",
		},
		{
			name: "shall return no errors given 'no' as input",
			val:  "no",
		},
		{
			name: "shall return no errors given empty string as input",
			val:  "",
		},
		{
			name: "shall return error given foo string as input",
			val:  "foo",
			wantErrs: []error{
				fmt.Errorf(`attribute %s does not support value %v
Supported values: 'yes', 'no', ''.`, key, "foo"),
			},
		},
		{
			name: "shall return error given 1 string as input",
			val:  "1",
			wantErrs: []error{
				fmt.Errorf(`attribute %s does not support value %v
Supported values: 'yes', 'no', ''.`, key, "1"),
			},
		},
		{
			name: "shall return error given true string as input",
			val:  "true",
			wantErrs: []error{
				fmt.Errorf(`attribute %s does not support value %v
Supported values: 'yes', 'no', ''.`, key, "true"),
			},
		},
		{
			name: "shall return error given true bool as input",
			val:  true,
			wantErrs: []error{
				fmt.Errorf(`attribute %s does not support value %v
Supported values: 'yes', 'no', ''.`, key, true),
			},
		},
	}

	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWarns, gotErrs := validateFuncNewOptionalTristateBool(tt.val, key)
			assert.Nilf(t, gotWarns, "unexpected slice of warnings")
			assert.Equalf(t, tt.wantErrs, gotErrs, "unexpected slice of errors")
		})
	}
}

func TestSetTristateBoolHappyPath(t *testing.T) {
	var (
		tV = true
		fV = false
	)
	type args struct {
		v *bool
	}
	const defaultAttrName = "attr_name"
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "shall set true",
			args: args{
				v: &tV,
			},
			want:    "yes",
			wantErr: false,
		},
		{
			name: "shall set false",
			args: args{
				v: &fV,
			},
			want:    "no",
			wantErr: false,
		},
		{
			name:    "shall set empty string",
			args:    args{},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := schema.Resource{
				Schema: map[string]*schema.Schema{
					defaultAttrName: NewOptionalTristateBool("", false),
				},
			}
			d := r.TestResourceData()
			if err := SetTristateBool(d, defaultAttrName, tt.args.v); (err != nil) != tt.wantErr {
				t.Error("unexpected error")
			}
			if !tt.wantErr {
				assert.Equal(t, tt.want, d.Get(defaultAttrName))
			}
		})
	}
}

func TestGetTristateBoolHappyPath(t *testing.T) {
	const defaultAttrName = "attr_name"
	r := schema.Resource{
		Schema: map[string]*schema.Schema{
			defaultAttrName: NewOptionalTristateBool("", false),
		},
	}
	var (
		tV = true
		fV = false
	)

	t.Parallel()

	tests := map[string]struct {
		setVal string
		want   *bool
	}{
		"shall read true":                    {setVal: "yes", want: &tV},
		"shall read false":                   {setVal: "no", want: &fV},
		"shall read false when attr not set": {setVal: "", want: nil},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d := r.TestResourceData()
			assert.NoError(t, d.Set(defaultAttrName, test.setVal))
			assert.Equal(t, test.want, GetTristateBool(d, defaultAttrName))
		})
	}
}
