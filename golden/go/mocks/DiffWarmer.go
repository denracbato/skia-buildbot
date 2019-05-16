// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import digest_counter "go.skia.org/infra/golden/go/digest_counter"
import digesttools "go.skia.org/infra/golden/go/digesttools"
import mock "github.com/stretchr/testify/mock"
import summary "go.skia.org/infra/golden/go/summary"
import types "go.skia.org/infra/golden/go/types"

// DiffWarmer is an autogenerated mock type for the DiffWarmer type
type DiffWarmer struct {
	mock.Mock
}

// PrecomputeDiffs provides a mock function with given fields: summaries, testNames, dCounter, diffFinder
func (_m *DiffWarmer) PrecomputeDiffs(summaries summary.SummaryMap, testNames types.TestNameSet, dCounter digest_counter.DigestCounter, diffFinder digesttools.ClosestDiffFinder) {
	_m.Called(summaries, testNames, dCounter, diffFinder)
}
