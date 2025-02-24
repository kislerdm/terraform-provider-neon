//go:build !acceptance
// +build !acceptance

package provider

import (
	"testing"
)

func Test_validateAutoscallingLimit(t *testing.T) {
	t.Parallel()

	t.Run(
		"happy path: int input", func(t *testing.T) {
			input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
			for _, in := range input {
				_, errs := validateAutoscallingLimit(in, "")
				if errs != nil {
					t.Fatal("errors are not expected")
				}
			}
		},
	)

	t.Run(
		"happy path: float64 input", func(t *testing.T) {
			input := []float64{0.25, 0.5, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
			for _, in := range input {
				_, errs := validateAutoscallingLimit(in, "")
				if errs != nil {
					t.Fatal("errors are not expected")
				}
			}
		},
	)

	t.Run(
		"unhappy path", func(t *testing.T) {
			input := []interface{}{"foo", 0, 0.1, 1.5, 11, 12, 20}
			for _, in := range input {
				_, errs := validateAutoscallingLimit(in, "")
				if errs == nil {
					t.Fatal("error is expected")
				}
			}
		},
	)
}
