package lcm_test

import (
	"testing"

	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
)

func TestProperties_Add(t *testing.T) {
	t.Parallel()

	type args struct {
		Name  string
		Value interface{}
	}

	tests := []struct {
		name string
		ps   lcm.Properties
		args args
		want int
	}{
		{
			name: "add_tc1",
			ps:   lcm.Properties{},
			args: args{
				Name:  "key",
				Value: "value",
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.ps.Add(tt.args.Name, tt.args.Value); len(tt.ps) != tt.want {
				t.Errorf("Add() = %v, want %v", len(tt.ps), tt.want)
			}
		})
	}
}
