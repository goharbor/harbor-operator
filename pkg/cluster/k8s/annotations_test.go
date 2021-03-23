package k8s

import (
	"testing"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
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
				o1: &goharborv1alpha2.Harbor{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							HarborClusterLastAppliedHash: "1",
						},
					},
					Spec: goharborv1alpha2.HarborSpec{},
				},
				o2: &goharborv1alpha2.Harbor{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							HarborClusterLastAppliedHash: "1",
						},
					},
					Spec: goharborv1alpha2.HarborSpec{},
				},
			},
			want: true,
		},
		{
			name: "TestHashEquals_2",
			args: args{
				o1: nil,
				o2: &goharborv1alpha2.Harbor{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							HarborClusterLastAppliedHash: "1",
						},
					},
					Spec: goharborv1alpha2.HarborSpec{},
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
				o1: &goharborv1alpha2.Harbor{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							HarborClusterLastAppliedHash: "1",
						},
					},
					Spec: goharborv1alpha2.HarborSpec{},
				},
				o2: &goharborv1alpha2.Harbor{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							HarborClusterLastAppliedHash: "2",
						},
					},
					Spec: goharborv1alpha2.HarborSpec{},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HashEquals(tt.args.o1, tt.args.o2); got != tt.want {
				t.Errorf("HashEquals() = %v, want %v", got, tt.want)
			}
		})
	}
}
