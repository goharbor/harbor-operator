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

package harborcluster

import (
	"context"

	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
)

// svcConfigGetter is used to get the required access data from the cluster spec for health checking.
type svcConfigGetter func(ctx context.Context, cluster *v1alpha2.HarborCluster) (*lcm.ServiceConfig, []lcm.Option, error)

// cacheConfigGetter is for getting configurations of cache service.
func cacheConfigGetter(ctx context.Context, cluster *v1alpha2.HarborCluster) (*lcm.ServiceConfig, []lcm.Option, error) {
	return nil, nil, nil
}

// dbConfigGetter is for getting configurations of database service.
func dbConfigGetter(ctx context.Context, cluster *v1alpha2.HarborCluster) (*lcm.ServiceConfig, []lcm.Option, error) {
	return nil, nil, nil
}

// storageConfigGetter is for getting configurations of storage service.
func storageConfigGetter(ctx context.Context, cluster *v1alpha2.HarborCluster) (*lcm.ServiceConfig, []lcm.Option, error) {
	return nil, nil, nil
}
