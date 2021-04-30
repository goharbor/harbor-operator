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

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/goharbor/harbor-operator/pkg/image"
	corev1 "k8s.io/api/core/v1"
)

const (
	ComponentName  = "cluster-minio"
	ConfigImageKey = "minio-docker-image"

	MinIOClientComponentName  = "cluster-minio-init"
	MinIOClientConfigImageKey = "minio-init-docker-image"
)

// getImage returns the configured image via configstore or default one.
func (m *MinIOController) getImage(ctx context.Context, harborcluster *goharborv1.HarborCluster) (string, error) {
	if harborcluster.Spec.InClusterStorage.MinIOSpec.Image != "" {
		return harborcluster.Spec.InClusterStorage.MinIOSpec.Image, nil
	}

	options := []image.Option{image.WithHarborVersion(harborcluster.Spec.Version)}
	if harborcluster.Spec.ImageSource != nil && (harborcluster.Spec.ImageSource.Repository != "" || harborcluster.Spec.ImageSource.TagSuffix != "") {
		options = append(options,
			image.WithRepository(harborcluster.Spec.ImageSource.Repository),
			image.WithTagSuffix(harborcluster.Spec.ImageSource.TagSuffix),
		)
	} else {
		options = append(options,
			image.WithConfigstore(m.ConfigStore),
			image.WithConfigImageKey(ConfigImageKey),
		)
	}

	image, err := image.GetImage(ctx, ComponentName, options...)
	if err != nil {
		return "", err
	}

	return image, nil
}

func (m *MinIOController) getImagePullPolicy(_ context.Context, harborcluster *goharborv1.HarborCluster) corev1.PullPolicy {
	if harborcluster.Spec.InClusterStorage.MinIOSpec.ImagePullPolicy != nil {
		return *harborcluster.Spec.InClusterStorage.MinIOSpec.ImagePullPolicy
	}

	if harborcluster.Spec.ImageSource != nil && harborcluster.Spec.ImageSource.ImagePullPolicy != nil {
		return *harborcluster.Spec.ImageSource.ImagePullPolicy
	}

	return config.DefaultImagePullPolicy
}

func (m *MinIOController) getImagePullSecret(_ context.Context, harborcluster *goharborv1.HarborCluster) corev1.LocalObjectReference {
	if len(harborcluster.Spec.InClusterStorage.MinIOSpec.ImagePullSecrets) > 0 {
		return harborcluster.Spec.InClusterStorage.MinIOSpec.ImagePullSecrets[0]
	}

	if harborcluster.Spec.ImageSource != nil && len(harborcluster.Spec.ImageSource.ImagePullSecrets) > 0 {
		return harborcluster.Spec.ImageSource.ImagePullSecrets[0]
	}

	return corev1.LocalObjectReference{Name: ""} // empty name means not using pull secret in minio
}

func (m *MinIOController) getMinIOClientImage(ctx context.Context, harborcluster *goharborv1.HarborCluster) (string, error) {
	spec := harborcluster.Spec.InClusterStorage.MinIOSpec.MinIOClientSpec
	if spec != nil && spec.Image != "" {
		return spec.Image, nil
	}

	options := []image.Option{image.WithHarborVersion(harborcluster.Spec.Version)}
	if harborcluster.Spec.ImageSource != nil && (harborcluster.Spec.ImageSource.Repository != "" || harborcluster.Spec.ImageSource.TagSuffix != "") {
		options = append(options,
			image.WithRepository(harborcluster.Spec.ImageSource.Repository),
			image.WithTagSuffix(harborcluster.Spec.ImageSource.TagSuffix),
		)
	} else {
		options = append(options,
			image.WithConfigstore(m.ConfigStore),
			image.WithConfigImageKey(MinIOClientConfigImageKey),
		)
	}

	image, err := image.GetImage(ctx, MinIOClientComponentName, options...)
	if err != nil {
		return "", err
	}

	return image, nil
}

func (m *MinIOController) getMinIOClientImagePullPolicy(_ context.Context, harborcluster *goharborv1.HarborCluster) corev1.PullPolicy {
	spec := harborcluster.Spec.InClusterStorage.MinIOSpec.MinIOClientSpec
	if spec != nil && spec.ImagePullPolicy != nil {
		return *spec.ImagePullPolicy
	}

	if harborcluster.Spec.ImageSource != nil && harborcluster.Spec.ImageSource.ImagePullPolicy != nil {
		return *harborcluster.Spec.ImageSource.ImagePullPolicy
	}

	return config.DefaultImagePullPolicy
}

func (m *MinIOController) getMinIOClientImagePullSecrets(_ context.Context, harborcluster *goharborv1.HarborCluster) []corev1.LocalObjectReference {
	spec := harborcluster.Spec.InClusterStorage.MinIOSpec.MinIOClientSpec
	if spec != nil && len(spec.ImagePullSecrets) > 0 {
		return spec.ImagePullSecrets
	}

	if harborcluster.Spec.ImageSource != nil && len(harborcluster.Spec.ImageSource.ImagePullSecrets) > 0 {
		return harborcluster.Spec.ImageSource.ImagePullSecrets
	}

	return nil
}
