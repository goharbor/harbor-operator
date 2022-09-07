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
	"fmt"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/harbor"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

// ServiceManager is designed to maintain the dependent services of the cluster.
type ServiceManager struct {
	ctx        context.Context
	cluster    *goharborv1.HarborCluster
	component  goharborv1.Component
	st         *status
	ctrl       lcm.Controller
	harborCtrl *harbor.Controller
}

// NewServiceManager constructs a new service manager for the specified component.
func NewServiceManager(component goharborv1.Component) *ServiceManager {
	return &ServiceManager{
		component: component,
	}
}

// WithContext bind a context.
func (s *ServiceManager) WithContext(ctx context.Context) *ServiceManager {
	s.ctx = ctx

	return s
}

// From which spec.
func (s *ServiceManager) From(cluster *goharborv1.HarborCluster) *ServiceManager {
	s.cluster = cluster

	return s
}

// For the harbor.
func (s *ServiceManager) For(harborCtrl *harbor.Controller) *ServiceManager {
	s.harborCtrl = harborCtrl

	return s
}

// TrackedBy by which status object.
func (s *ServiceManager) TrackedBy(st *status) *ServiceManager {
	s.st = st

	return s
}

// Use which ctrl.
func (s *ServiceManager) Use(ctrl lcm.Controller) *ServiceManager {
	s.ctrl = ctrl

	return s
}

// Apply changes.
func (s *ServiceManager) Apply() error { //nolint:funlen
	if err := s.validate(); err != nil {
		return err
	}

	var (
		status        *lcm.CRStatus
		err           error
		conditionType = s.conditionType()
	)

	defer func() {
		// Add condition
		if status != nil {
			s.st.log.Info(fmt.Sprintf("%s readiness is %s, %s, %s", s.component, status.Condition.Status, status.Condition.Reason, status.Condition.Message))
			// Here just add condition update to the status, the real update happens outside
			s.st.UpdateCondition(conditionType, status.Condition)
			// Assign to harbor ctrl for the reconcile of cluster
			if err == nil {
				s.st.TrackDependencies(s.component, status)
			}
		}
	}()

	// The validating webhook validates the spec and either incluster or external can be configured.
	useInCluster := true

	switch s.component {
	case goharborv1.ComponentCache:
		if s.cluster.Spec.Cache.Spec.RedisFailover == nil {
			useInCluster = false
		}
	case goharborv1.ComponentDatabase:
		if s.cluster.Spec.Database.Spec.ZlandoPostgreSQL == nil {
			useInCluster = false
		}
	case goharborv1.ComponentStorage:
		if s.cluster.Spec.Storage.Spec.MinIO == nil {
			useInCluster = false
		}
		// Only for wsl check
	case goharborv1.ComponentHarbor:
		return errors.Errorf("%s is not supported", s.component)
	default:
		// Should not happen, just in case
		return errors.Errorf("unrecognized component: %s", s.component)
	}

	if s.ctx == nil {
		s.ctx = context.TODO()
	}

	if useInCluster {
		// Use incluster
		status, err = s.ctrl.Apply(s.ctx, s.cluster)
		if err != nil {
			return err
		}
	} else {
		// Default is ready
		status = &lcm.CRStatus{
			Condition: goharborv1.HarborClusterCondition{
				Type:   conditionType,
				Status: v1.ConditionTrue,
			},
		}
	}

	return nil
}

func (s *ServiceManager) validate() error {
	if s.component != goharborv1.ComponentCache &&
		s.component != goharborv1.ComponentStorage &&
		s.component != goharborv1.ComponentDatabase {
		return errors.Errorf("invalid service component: %s", s.component)
	}

	if s.cluster == nil {
		return errors.New("missing cluster spec")
	}

	if s.ctrl == nil {
		return errors.New("missing lcm controller")
	}

	if s.st == nil {
		return errors.New("missing status")
	}

	if s.harborCtrl == nil {
		return errors.New("missing harbor ctrl")
	}

	return nil
}

func (s *ServiceManager) conditionType() goharborv1.HarborClusterConditionType {
	switch s.component {
	case goharborv1.ComponentStorage:
		return goharborv1.StorageReady
	case goharborv1.ComponentDatabase:
		return goharborv1.DatabaseReady
	case goharborv1.ComponentCache:
		return goharborv1.CacheReady
		// Only for wsl check
	case goharborv1.ComponentHarbor:
		return goharborv1.ServiceReady
	default:
		// Should not reach here
		return ""
	}
}
