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

package gos

import (
	"context"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// Group for running goroutines.
type Group struct {
	wait   sync.WaitGroup
	cancel context.CancelFunc

	// Collect all the errors
	errors []error
	locker *sync.Mutex
}

// NewGroup creates a new group.
func NewGroup(ctx context.Context) (*Group, context.Context) {
	gtx, cancel := context.WithCancel(ctx)

	return &Group{
		cancel: cancel,
		errors: make([]error, 0),
		locker: &sync.Mutex{},
	}, gtx
}

// Go to run a func.
func (g *Group) Go(f func() error) {
	g.wait.Add(1)

	go func() {
		defer g.wait.Done()

		// Run and handle error
		if err := f(); err != nil {
			g.locker.Lock()
			defer g.locker.Unlock()

			g.errors = append(g.errors, err)
		}
	}()
}

// Wait until all the go functions are returned
// If errors occurred, they'll be combined with ":" and returned.
func (g *Group) Wait() error {
	g.wait.Wait()

	defer func() {
		if g.cancel != nil {
			// Send signals to the downstream components
			g.cancel()
		}
	}()

	errTexts := make([]string, 0)
	for _, e := range g.errors {
		errTexts = append(errTexts, e.Error())
	}

	if len(errTexts) > 0 {
		return errors.Errorf("gos.Group error: %s", strings.Join(errTexts, ":"))
	}

	return nil
}
