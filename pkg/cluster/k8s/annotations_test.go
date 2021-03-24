package k8s_test

import (
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
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

func TestSetLastAppliedHash(t *testing.T) {
	HarborCR1 := &goharborv1.Harbor{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: goharborv1.HarborSpec{
			LogLevel: "info",
			Version:  "2.12",
		},
		Status: harbormetav1.ComponentStatus{},
	}
	HarborCR2 := &goharborv1.Harbor{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: goharborv1.HarborSpec{
			LogLevel: "info",
			Version:  "2.12",
		},
		Status: harbormetav1.ComponentStatus{},
	}

	HarborCR3 := &goharborv1.Harbor{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: goharborv1.HarborSpec{
			LogLevel: "info",
			Version:  "2.12",
		},
		Status: harbormetav1.ComponentStatus{
			ObservedGeneration: 1,
			Replicas:           func(v int32) *int32 { return &v }(2),
		},
	}

	HarborCR4 := &goharborv1.Harbor{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: goharborv1.HarborSpec{
			LogLevel: "debug",
			Version:  "2.12",
		},
		Status: harbormetav1.ComponentStatus{
			ObservedGeneration: 1,
			Replicas:           func(v int32) *int32 { return &v }(2),
		},
	}

	k8s.SetLastAppliedHash(HarborCR1, HarborCR1.Spec)
	k8s.SetLastAppliedHash(HarborCR2, HarborCR2.Spec)
	k8s.SetLastAppliedHash(HarborCR3, HarborCR3.Spec)
	k8s.SetLastAppliedHash(HarborCR4, HarborCR4.Spec)

	if !k8s.HashEquals(HarborCR1, HarborCR2) {
		t.Error("Expect HashEquals true, but false")
	}

	if !k8s.HashEquals(HarborCR1, HarborCR3) {
		t.Error("Expect HashEquals true, but false")
	}

	if k8s.HashEquals(HarborCR1, HarborCR4) {
		t.Error("Expect HashEquals false, but true")
	}
}
