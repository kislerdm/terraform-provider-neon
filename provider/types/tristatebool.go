package types

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	ValTrue  = "yes"
	ValFalse = "no"
	ValNull  = ""
)

func validateFuncNewOptionalTristateBool(v interface{}, s string) (warns []string, errs []error) {
	const supportedVals = "Supported values: '" + ValTrue + "', '" +
		ValFalse + "', '" + ValNull + "'."

	vv, ok := v.(string)
	if ok {
		switch vv {
		case ValTrue, ValFalse, ValNull:
		default:
			ok = false
		}
	}

	if !ok {
		errs = []error{
			fmt.Errorf("attribute %s does not support value %v\n%s", s, v, supportedVals),
		}
	}

	return warns, errs
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
	var err error
	switch {
	case v == nil:
		err = d.Set(name, ValNull)
	case *v:
		err = d.Set(name, ValTrue)
	default:
		err = d.Set(name, ValFalse)
	}
	return err
}

// GetTristateBool reads the bool value from the tristate bool value of the resource's definition
// using the attribute vlue.
func GetTristateBool(d *schema.ResourceData, name string) *bool {
	var o *bool = nil
	switch d.Get(name) {
	case ValNull:
	case ValFalse:
		var tmp bool
		o = &tmp
	case ValTrue:
		tmp := true
		o = &tmp
	}
	return o
}

// IsNull checks if the attribute's value is null.
func IsNull(d *schema.ResourceData, name string) bool {
	return d.Get(name) == ValNull
}
