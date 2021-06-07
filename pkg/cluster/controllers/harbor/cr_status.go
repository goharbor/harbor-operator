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

package harbor

import (
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	corev1 "k8s.io/api/core/v1"
)

// harborReadyStatus indicates harbor (CR) is ready.
var harborReadyStatus = lcm.New(goharborv1.ServiceReady).WithStatus(corev1.ConditionTrue)

// harborNotReadyStatus indicates harbor (CR) is not ready.
var harborNotReadyStatus = func(reason, message string) *lcm.CRStatus {
	return lcm.New(goharborv1.ServiceReady).WithStatus(corev1.ConditionFalse).WithReason(reason).WithMessage(message)
}

// harborUnknownStatus indicates status of harbor (CR) is unknown.
var harborUnknownStatus = func(reason, message string) *lcm.CRStatus {
	return lcm.New(goharborv1.ServiceReady).WithStatus(corev1.ConditionUnknown).WithReason(reason).WithMessage(message)
}
