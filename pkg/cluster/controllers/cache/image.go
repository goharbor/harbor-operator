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

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/image"
)

const (
	ComponentName  = "cluster-redis"
	ConfigImageKey = "redis-docker-image"
)

// GetImage returns the configured image via configstore or default one.
func (rm *redisResourceManager) GetImage(ctx context.Context, harborcluster *goharborv1alpha2.HarborCluster) (string, error) {
	if harborcluster.Spec.InClusterCache.RedisSpec.Image != "" {
		return harborcluster.Spec.InClusterCache.RedisSpec.Image, nil
	}

	options := []image.Option{image.WithHarborVersion(harborcluster.Spec.Version)}
	if harborcluster.Spec.ImageSource != nil && (harborcluster.Spec.ImageSource.Repository != "" || harborcluster.Spec.ImageSource.TagSuffix != "") {
		options = append(options,
			image.WithRepository(harborcluster.Spec.ImageSource.Repository),
			image.WithTagSuffix(harborcluster.Spec.ImageSource.TagSuffix),
		)
	} else {
		options = append(options,
			image.WithConfigstore(rm.configStore),
			image.WithConfigImageKey(ConfigImageKey),
		)
	}

	image, err := image.GetImage(ctx, ComponentName, options...)
	if err != nil {
		return "", err
	}

	return image, nil
}
