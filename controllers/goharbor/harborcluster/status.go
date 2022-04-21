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
	"sync"
	"time"

	"github.com/go-logr/logr"
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// status is designed to track the status and conditions of the deploying Harbor cluster.
type status struct {
	client.Client
	log     *logr.Logger
	context context.Context

	cr             *goharborv1.HarborCluster
	data           *goharborv1.HarborClusterStatus
	sourceRevision int64
	collection     *lcm.CRStatusCollection

	locker *sync.Mutex
}

// NewStatus constructs a new status.
func newStatus(source *goharborv1.HarborCluster) *status {
	// New with default status and conditions
	s := &status{
		cr:     source,
		locker: &sync.Mutex{},
		data: &goharborv1.HarborClusterStatus{
			Status:     goharborv1.StatusProvisioning,
			Revision:   time.Now().UnixNano(),
			Conditions: make([]goharborv1.HarborClusterCondition, 0),
		},
		collection: lcm.NewCRStatusCollection(),
	}

	if source != nil {
		s.data.Operator = source.Status.Operator
	}

	// Copy source status if it has been set before
	if source != nil && len(source.Status.Status) > 0 {
		s.data.Status = source.Status.Status
		s.data.Revision = source.Status.Revision
		s.data.Conditions = append(s.data.Conditions, source.Status.Conditions...)
		s.sourceRevision = source.Status.Revision // for comparison later
	}

	return s
}

// Update the status.
func (s *status) Update() error {
	// In case
	s.locker.Lock()
	defer s.locker.Unlock()

	// If we need to do the status update
	if s.sourceRevision == s.data.Revision {
		// do nothing
		return nil
	}

	s.log.Info("status revision changed", "original", s.sourceRevision, "current", s.data.Revision)

	// Validate client
	if err := s.validate(); err != nil {
		return err
	}

	// Override status
	s.data.Status = s.overallStatus()
	s.cr.Status = *s.data

	if err := s.Client.Status().Update(s.context, s.cr); err != nil {
		if apierrors.IsConflict(err) {
			s.log.Error(err, "failed to update status of harbor cluster")

			return nil
		}

		return err
	}

	return nil
}

// DependsReady judges if all the dependent services are ready.
func (s *status) DependsReady() bool {
	// In case
	s.locker.Lock()
	defer s.locker.Unlock()

	for _, c := range s.data.Conditions {
		if c.Type == goharborv1.CacheReady ||
			c.Type == goharborv1.DatabaseReady ||
			c.Type == goharborv1.StorageReady {
			if c.Status != corev1.ConditionTrue {
				return false
			}
		}
	}

	return true
}

// For the harbor cluster CR.
func (s *status) For(resource *goharborv1.HarborCluster) *status {
	s.cr = resource

	return s
}

// WithClient set client.
func (s *status) WithClient(c client.Client) *status {
	s.Client = c

	return s
}

// WithContext set context.
func (s *status) WithContext(ctx context.Context) *status {
	s.context = ctx

	return s
}

// WithLog set logger.
func (s *status) WithLog(logger logr.Logger) *status {
	s.log = &logger

	return s
}

// UpdateCondition adds condition update of the specified service to the status object.
func (s *status) UpdateCondition(ct goharborv1.HarborClusterConditionType, c goharborv1.HarborClusterCondition) {
	s.locker.Lock()
	defer s.locker.Unlock()

	for i := range s.data.Conditions {
		cp := &s.data.Conditions[i]

		if cp.Type == ct {
			if cp.Status != c.Status ||
				cp.Reason != c.Reason ||
				cp.Message != c.Message {
				// Override
				cp.Status = c.Status
				cp.Message = c.Message
				cp.Reason = c.Reason
				// Update timestamp
				now := metav1.Now()
				cp.LastTransitionTime = &now

				// Update revision for identifying the changes
				s.data.Revision = time.Now().UnixNano()
			}

			return
		}
	}
	// Append if not existing yet
	cc := c.DeepCopy()
	now := metav1.Now()
	cc.LastTransitionTime = &now
	s.data.Conditions = append(s.data.Conditions, *cc)
	s.data.Revision = time.Now().UnixNano()
}

// TrackDependencies tracks the status of the dependant CR.
func (s *status) TrackDependencies(component goharborv1.Component, crStatus *lcm.CRStatus) {
	s.collection.Set(component, crStatus)
}

// GetDependencies returns the status collection.
func (s *status) GetDependencies() *lcm.CRStatusCollection {
	return s.collection
}

func (s *status) overallStatus() goharborv1.ClusterStatus {
	var ready, unready int

	for _, c := range s.data.Conditions {
		if !isDependencyConditionType(c.Type) {
			continue
		}

		switch c.Status {
		case corev1.ConditionTrue:
			ready++
		case corev1.ConditionFalse:
			unready++
		case corev1.ConditionUnknown:
		default:
		}
	}

	// Totally ready
	const TotalDependencyNum = 4
	if ready >= TotalDependencyNum {
		return goharborv1.StatusHealthy
	}

	// Any related components are unhealthy, cluster should be marked as unhealthy
	if unready >= 1 {
		return goharborv1.StatusUnHealthy
	}

	return goharborv1.StatusProvisioning
}

func (s *status) validate() error {
	if s.cr == nil {
		return errors.New("missing harbor cluster CR")
	}

	if s.Client == nil {
		return errors.New("client is not set")
	}

	if s.context == nil {
		return errors.New("missing context")
	}

	if s.log == nil {
		return errors.New("missing logger")
	}

	return nil
}

func isDependencyConditionType(t goharborv1.HarborClusterConditionType) bool {
	return t == goharborv1.ServiceReady || t == goharborv1.CacheReady ||
		t == goharborv1.DatabaseReady || t == goharborv1.StorageReady
}
