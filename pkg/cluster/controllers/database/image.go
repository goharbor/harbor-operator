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

const (
	DefaultCacheImage = "registry.opensource.zalan.do/acid/spilo-12:1.6-p3"
	ConfigImageKey    = "database-image"
)

// GetImage returns the configured image via configstore or default one.
func (p *PostgreSQLController) GetImage() string {
	image, err := p.ConfigStore.GetItemValue(ConfigImageKey)
	if err != nil {
		// Just logged
		p.Log.V(5).Error(err, "get database image error", "image-key", ConfigImageKey)

		image = DefaultCacheImage
	}

	return image
}
