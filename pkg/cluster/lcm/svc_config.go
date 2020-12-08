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

package lcm

import (
	"context"

	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SvcConfigGetter is used to get the required access data from the cluster spec for health checking.
type SvcConfigGetter interface {
	WithCtx(ctx context.Context) SvcConfigGetter
	UseClient(client client.Client) SvcConfigGetter
	FromCluster(cluster *v1alpha2.HarborCluster) SvcConfigGetter
	GetConfig() (*ServiceConfig, []Option, error)
}
