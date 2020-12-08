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
	"errors"

	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/goharbor/harbor-operator/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// svcConfigGetter is used to get the required access data from the cluster spec for health checking.
type svcConfigGetter func(ctx context.Context, kubeClient k8s.Client, cluster *v1alpha2.HarborCluster) (*lcm.ServiceConfig, []lcm.Option, error)

// cacheConfigGetter is for getting configurations of cache service.
func cacheConfigGetter(ctx context.Context, kubeClient k8s.Client, cluster *v1alpha2.HarborCluster) (*lcm.ServiceConfig, []lcm.Option, error) {
	return nil, nil, nil
}

// dbConfigGetter is for getting configurations of database service.
func dbConfigGetter(ctx context.Context, kubeClient k8s.Client, cluster *v1alpha2.HarborCluster) (*lcm.ServiceConfig, []lcm.Option, error) {
	var (
		host, accessKey, accessSecret, secret string
		port                                  uint
		option                                []lcm.Option
	)

	if cluster.Spec.HarborComponentsSpec.Database == nil {
		return nil, nil, errors.New("cluster.Spec.HarborComponentsSpec.Database invalid value")
	}

	db := cluster.Spec.HarborComponentsSpec.Database

	host = db.Hosts[0].Host
	if len(host) == 0 {
		return nil, nil, errors.New("Database.Hosts invalid value")
	}

	port = uint(int(db.Hosts[0].Port))
	if port == 0 {
		return nil, nil, errors.New("Database.Port invalid value")
	}

	accessKey = db.Username
	if len(accessKey) == 0 {
		return nil, nil, errors.New("Database.Username invalid value")
	}

	secret = db.PasswordRef
	if len(secret) == 0 {
		return nil, nil, errors.New("Database.PasswordRef invalid value")
	}

	accessSecret, err := getAccessSecret(kubeClient, secret, cluster.GetNamespace())
	if err != nil {
		return nil, nil, err
	}

	return &lcm.ServiceConfig{
		Endpoint: &lcm.Endpoint{
			Host: db.Hosts[0].Host,
			Port: port,
		},
		Credentials: &lcm.Credentials{
			AccessKey:    accessKey,
			AccessSecret: accessSecret,
		},
	}, option, nil
}

// storageConfigGetter is for getting configurations of storage service.
func storageConfigGetter(ctx context.Context, kubeClient k8s.Client, cluster *v1alpha2.HarborCluster) (*lcm.ServiceConfig, []lcm.Option, error) {
	return nil, nil, nil
}

// getAccessSecret is for getting component connection password.
func getAccessSecret(kubeClient k8s.Client, name, ns string) (string, error) {
	var accessSecret string

	secret, err := getSecret(kubeClient, name, ns)
	if err != nil {
		return accessSecret, err
	}

	for k, v := range secret {
		switch k {
		case harbormetav1.PostgresqlPasswordKey:
			accessSecret = string(v)
		case harbormetav1.RedisPasswordKey:
			accessSecret = string(v)
		}
	}

	return accessSecret, nil
}

// getSecret is for getting secret.
func getSecret(kubeClient k8s.Client, secretName, ns string) (map[string][]byte, error) {
	secret := &corev1.Secret{}

	if err := kubeClient.Get(types.NamespacedName{Name: secretName, Namespace: ns}, secret); err != nil {
		return nil, err
	}

	return secret.Data, nil
}
