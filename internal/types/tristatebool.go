package types

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	valTrue  = "yes"
	valFalse = "no"
	valNull  = ""
)

func validateFuncNewOptionalTristateBool(v interface{}, s string) (warns []string, errs []error) {
	switch vv := v.(string); vv {
	case valTrue, valFalse, valNull:
		return
	default:
		const supportedVals = "Supported values: '" + valTrue + "', '" +
			valFalse + "', '" + valNull + "'."
		return nil, []error{
			errors.New("attribute " + s + " does not support value " + vv + "\n" + supportedVals),
		}
	}
}

// NewOptionalTristateBool initialises the tristate bool value.
// See discussion: https://github.com/hashicorp/terraform-plugin-sdk/issues/817
func NewOptionalTristateBool(description string, forceNew bool) *schema.Schema {
	const descriptionHeader = "Set to 'yes' to activate, 'no' to deactivate explicitly, and omit to keep the default value.\n"
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     forceNew,
		ValidateFunc: validateFuncNewOptionalTristateBool,
		Description:  descriptionHeader + description,
	}
}

// SetTristateBool sets the tristate bool value.
// The Adapter to the schema.ResourceData{}.Set method to convert pointer to bool to
// the string equivalent of bool (yes/no) to maintain the tristate bool.
func SetTristateBool(d *schema.ResourceData, name string, v *bool) error {
	var setValue string
	switch {
	case v == nil:
	case *v:
		setValue = valTrue
	default:
		setValue = valFalse
	}
	return d.Set(name, setValue)
}

// GetTristateBool reads the bool value from the tristate bool value of the resource's definition
// using the attribute vlue.
func GetTristateBool(d *schema.ResourceData, name string) *bool {
	var o *bool = nil
	switch d.Get(name) {
	case valNull:
	case valFalse:
		var tmp bool
		o = &tmp
	case valTrue:
		tmp := true
		o = &tmp
	}
	return o
}
