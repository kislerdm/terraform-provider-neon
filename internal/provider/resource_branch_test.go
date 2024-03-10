package provider

import (
	"os"
	"testing"
)

func Test_isValidBranchID(t *testing.T) {
	if os.Getenv("TF_ACC") == "1" {
		t.Skip("acceptance tests are running")
	}

	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "happy path",
			args: args{
				s: "br-foo123b",
			},
			want: true,
		},
		{
			name: "unhappy path: no id post-prefix",
			args: args{
				s: "br-",
			},
			want: false,
		},
		{
			name: "unhappy path: no prefix",
			args: args{
				s: "qux",
			},
			want: false,
		},
		{
			name: "unhappy path: empty string",
			args: args{
				s: "",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidBranchID(tt.args.s); got != tt.want {
				t.Errorf("isValidBranchID() = %v, want %v", got, tt.want)
			}
		})
	}
}
