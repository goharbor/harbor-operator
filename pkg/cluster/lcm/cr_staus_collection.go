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
	"sync"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
)

// CRStatusCollection is designed for collecting CRStatus of each dependant components.
type CRStatusCollection struct {
	componentToCRStatus *sync.Map
}

// NewCRStatusCollection returns a new collection.
func NewCRStatusCollection() *CRStatusCollection {
	return &CRStatusCollection{
		componentToCRStatus: &sync.Map{},
	}
}

// Set item to collection.
func (c *CRStatusCollection) Set(component goharborv1.Component, status *CRStatus) {
	if component != "" && status != nil {
		c.componentToCRStatus.Store(component, status)
	}
}

// Get item from collection.
func (c *CRStatusCollection) Get(component goharborv1.Component) (*CRStatus, bool) {
	v, ok := c.componentToCRStatus.Load(component)
	if ok {
		return v.(*CRStatus), ok
	}

	return nil, false
}
