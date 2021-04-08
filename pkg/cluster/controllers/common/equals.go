package common

import (
	"context"

	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"github.com/goharbor/harbor-operator/pkg/resources/checksum"
	"k8s.io/apimachinery/pkg/runtime"
)

func Equals(ctx context.Context, s *runtime.Scheme, cluster *v1alpha3.HarborCluster, obj checksum.Dependency) bool {
	dependency := checksum.New(s)
	dependency.Add(ctx, cluster, true)

	return !dependency.ChangedFor(ctx, obj)
}
