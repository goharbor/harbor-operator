// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cache

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/image"
)

const (
	ComponentName = "cluster-redis"
)

// GetImage returns the configured image via configstore or default one.
func (rm *redisResourceManager) GetImage(ctx context.Context, harborcluster *goharborv1.HarborCluster) (string, error) {
	options := harborcluster.Spec.ImageSource.AddRepositoryAndTagSuffixOptions(
		image.WithImageFromSpec(harborcluster.Spec.Cache.Spec.RedisFailover.Image),
		image.WithHarborVersion(harborcluster.Spec.Version),
	)

	return image.GetImage(ctx, ComponentName, options...)
}
