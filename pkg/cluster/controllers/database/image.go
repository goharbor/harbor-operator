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

package database

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/image"
)

const (
	ComponentName  = "cluster-postgresql"
	ConfigImageKey = "postgresql-docker-image"
)

// GetImage returns the configured image via configstore or default one.
func (p *PostgreSQLController) GetImage(ctx context.Context, harborcluster *goharborv1.HarborCluster) (string, error) {
	if harborcluster.Spec.Database.Spec.ZlandoPostgreSQL.Image != "" {
		return harborcluster.Spec.Database.Spec.ZlandoPostgreSQL.Image, nil
	}

	options := []image.Option{image.WithHarborVersion(harborcluster.Spec.Version)}
	if harborcluster.Spec.ImageSource != nil && (harborcluster.Spec.ImageSource.Repository != "" || harborcluster.Spec.ImageSource.TagSuffix != "") {
		options = append(options,
			image.WithRepository(harborcluster.Spec.ImageSource.Repository),
			image.WithTagSuffix(harborcluster.Spec.ImageSource.TagSuffix),
		)
	} else {
		options = append(options,
			image.WithConfigstore(p.ConfigStore),
			image.WithConfigImageKey(ConfigImageKey),
		)
	}

	return image.GetImage(ctx, ComponentName, options...)
}
