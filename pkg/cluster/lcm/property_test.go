package lcm

import (
	"fmt"
	"strings"
	"testing"
)

func TestProperties_Add(t *testing.T) {
	type args struct {
		Name  string
		Value interface{}
	}
	tests := []struct {
		name string
		ps   Properties
		args args
		want int
	}{
		{
			name: "add_tc1",
			ps:   Properties{},
			args: args{
				Name:  "key",
				Value: "value",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ps.Add(tt.args.Name, tt.args.Value); len(tt.ps) != tt.want {
				t.Errorf("Add() = %v, want %v", len(tt.ps), tt.want)
			}
		})
	}
}

func Test(t *testing.T) {
	component := "chartMuseum"
	name := strings.ToLower(fmt.Sprintf("%s-redis", component))
	fmt.Println(strings.ToLower(name))
}
