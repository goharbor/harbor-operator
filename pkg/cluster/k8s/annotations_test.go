package k8s_test

import (
	"testing"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"github.com/goharbor/harbor-operator/pkg/cluster/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHashEquals(t *testing.T) {
	type args struct {
		o1 metav1.Object
		o2 metav1.Object
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "TestHashEquals_1",
			args: args{
				o1: &goharborv1.Harbor{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							k8s.HarborClusterLastAppliedHash: "1",
						},
					},
					Spec: goharborv1.HarborSpec{},
				},
				o2: &goharborv1.Harbor{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							k8s.HarborClusterLastAppliedHash: "1",
						},
					},
					Spec: goharborv1.HarborSpec{},
				},
			},
			want: true,
		},
		{
			name: "TestHashEquals_2",
			args: args{
				o1: nil,
				o2: &goharborv1.Harbor{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							k8s.HarborClusterLastAppliedHash: "1",
						},
					},
					Spec: goharborv1.HarborSpec{},
				},
			},
			want: false,
		},
		{
			name: "TestHashEquals_3",
			args: args{
				o1: nil,
				o2: nil,
			},
			want: true,
		},
		{
			name: "TestHashEquals_4",
			args: args{
				o1: &goharborv1.Harbor{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							k8s.HarborClusterLastAppliedHash: "1",
						},
					},
					Spec: goharborv1.HarborSpec{},
				},
				o2: &goharborv1.Harbor{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							k8s.HarborClusterLastAppliedHash: "2",
						},
					},
					Spec: goharborv1.HarborSpec{},
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := k8s.HashEquals(tc.args.o1, tc.args.o2); got != tc.want {
				t.Errorf("HashEquals() = %v, want %v", got, tc.want)
			}
		})
	}
}
