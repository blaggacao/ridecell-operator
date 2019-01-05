/*
Copyright 2019 Ridecell, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package matchers

import (
	"fmt"

	"github.com/onsi/gomega/types"
	"github.com/pkg/errors"

	"github.com/Ridecell/ridecell-operator/pkg/components"
)

// Gomega matcher for checking if a component reconciles correctly.
//
//   comp := MyComponent()
//   Expect(comp).To(ReconcileContext(ctx))
func ReconcileContext(ctx *components.ComponentContext) types.GomegaMatcher {
	return &reconcileContextMatcher{ctx: ctx}
}

type reconcileContextMatcher struct {
	ctx *components.ComponentContext
  err error
}

// Match implements GomegaMatcher
func (matcher *reconcileContextMatcher) Match(actual interface{}) (bool, error) {
	// Check interface compliance.
	comp, ok := actual.(components.Component)
	if !ok {
		return false, errors.Errorf("ReconcileContext matcher expects a ComponentContext")
	}

	// Run the reconcile.
	result, err := comp.Reconcile(matcher.ctx)
	if err != nil {
    matcher.err = err
		return false, nil
	}

	// Check if we have a StatusModifier.
	if result.StatusModifier != nil {
		err = result.StatusModifier(matcher.ctx.Top)
		if err != nil {
      matcher.err = err
			return false, nil
		}
	}

	return true, nil
}

// FailureMessage implements GomegaMatcher
func (matcher *reconcileContextMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto reconcile succesfully, got \n\t%s", actual, matcher.err)
}

// NegatedFailureMessage implements GomegaMatcher
func (matcher *reconcileContextMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto error during reconcile, no error received", actual)
}
