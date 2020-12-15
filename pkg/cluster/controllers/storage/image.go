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

package storage

import "fmt"

const (
	DefaultCacheImage = "minio/minio:RELEASE.2020-08-13T02-39-50Z"
	ConfigImageKey    = "storage-image"
)

// GetImage returns the configured image via configstore or default one.
func (m *MinIOController) GetImage() string {
	if m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Version != "" {
		return fmt.Sprintf("minio/minio:%s", m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Version)
	}

	image, err := m.ConfigStore.GetItemValue(ConfigImageKey)
	if err != nil {
		// Just logged
		m.Log.V(5).Error(err, "get storage image error", "image-key", ConfigImageKey)

		image = DefaultCacheImage
	}

	return image
}
