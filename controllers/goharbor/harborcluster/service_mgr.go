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
	"fmt"

	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/harbor"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	v1 "k8s.io/api/core/v1"
)

// ServiceManager is designed to maintain the dependent services of the cluster.
type ServiceManager struct {
	ctx             context.Context
	cluster         *v1alpha2.HarborCluster
	component       v1alpha2.Component
	st              *status
	ctrl            lcm.Controller
	svcConfigGetter svcConfigGetter
	harborCtrl      *harbor.Controller
}

// NewServiceManager constructs a new service manager for the specified component.
func NewServiceManager(component v1alpha2.Component) *ServiceManager {
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
func (s *ServiceManager) From(cluster *v1alpha2.HarborCluster) *ServiceManager {
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

// WithConfig bind service configuration getter func.
func (s *ServiceManager) WithConfig(svcCfgGetter svcConfigGetter) *ServiceManager {
	s.svcConfigGetter = svcCfgGetter

	return s
}

// Apply changes.
// nolint:funlen
func (s *ServiceManager) Apply() error {
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
			// Here just add condition update to the status, the real update happens outside
			s.st.UpdateCondition(conditionType, status.Condition)
			// Assign to harbor ctrl for the reconcile of cluster
			if err == nil {
				s.harborCtrl.WithDependency(s.component, status)
			}
		}
	}()

	// The validating webhook validates the spec and either incluster or external can be configured.
	useInCluster := true

	switch s.component {
	case v1alpha2.ComponentCache:
		if s.cluster.Spec.InClusterCache == nil {
			useInCluster = false
		}
	case v1alpha2.ComponentDatabase:
		if s.cluster.Spec.InClusterDatabase == nil {
			useInCluster = false
		}
	case v1alpha2.ComponentStorage:
		if s.cluster.Spec.InClusterStorage == nil {
			useInCluster = false
		}
		// Only for wsl check
	case v1alpha2.ComponentHarbor:
		return fmt.Errorf("%s is not supported", s.component)
	default:
		// Should not happen, just in case
		return fmt.Errorf("unrecognized component: %s", s.component)
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
			Condition: v1alpha2.HarborClusterCondition{
				Type:   conditionType,
				Status: v1.ConditionTrue,
			},
		}
	}

	return nil
}

func (s *ServiceManager) validate() error {
	if s.component != v1alpha2.ComponentCache &&
		s.component != v1alpha2.ComponentStorage &&
		s.component != v1alpha2.ComponentDatabase {
		return fmt.Errorf("invalid service component: %s", s.component)
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

	if s.svcConfigGetter == nil {
		return errors.New("missing svc config getter")
	}

	if s.harborCtrl == nil {
		return errors.New("missing harbor ctrl")
	}

	return nil
}

func (s *ServiceManager) conditionType() v1alpha2.HarborClusterConditionType {
	switch s.component {
	case v1alpha2.ComponentStorage:
		return v1alpha2.StorageReady
	case v1alpha2.ComponentDatabase:
		return v1alpha2.DatabaseReady
	case v1alpha2.ComponentCache:
		return v1alpha2.CacheReady
		// Only for wsl check
	case v1alpha2.ComponentHarbor:
		return v1alpha2.ServiceReady
	default:
		// Should not reach here
		return ""
	}
}
