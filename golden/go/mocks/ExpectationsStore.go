// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	context "context"

	expstorage "go.skia.org/infra/golden/go/expstorage"
	expectations "go.skia.org/infra/golden/go/types/expectations"

	mock "github.com/stretchr/testify/mock"
)

// ExpectationsStore is an autogenerated mock type for the ExpectationsStore type
type ExpectationsStore struct {
	mock.Mock
}

// AddChange provides a mock function with given fields: ctx, changes, userId
func (_m *ExpectationsStore) AddChange(ctx context.Context, changes expectations.Expectations, userId string) error {
	ret := _m.Called(ctx, changes, userId)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, expectations.Expectations, string) error); ok {
		r0 = rf(ctx, changes, userId)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ForChangeList provides a mock function with given fields: id, crs
func (_m *ExpectationsStore) ForChangeList(id string, crs string) expstorage.ExpectationsStore {
	ret := _m.Called(id, crs)

	var r0 expstorage.ExpectationsStore
	if rf, ok := ret.Get(0).(func(string, string) expstorage.ExpectationsStore); ok {
		r0 = rf(id, crs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(expstorage.ExpectationsStore)
		}
	}

	return r0
}

// Get provides a mock function with given fields:
func (_m *ExpectationsStore) Get() (expectations.Expectations, error) {
	ret := _m.Called()

	var r0 expectations.Expectations
	if rf, ok := ret.Get(0).(func() expectations.Expectations); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(expectations.Expectations)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// QueryLog provides a mock function with given fields: ctx, offset, n, details
func (_m *ExpectationsStore) QueryLog(ctx context.Context, offset int, n int, details bool) ([]expstorage.TriageLogEntry, int, error) {
	ret := _m.Called(ctx, offset, n, details)

	var r0 []expstorage.TriageLogEntry
	if rf, ok := ret.Get(0).(func(context.Context, int, int, bool) []expstorage.TriageLogEntry); ok {
		r0 = rf(ctx, offset, n, details)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]expstorage.TriageLogEntry)
		}
	}

	var r1 int
	if rf, ok := ret.Get(1).(func(context.Context, int, int, bool) int); ok {
		r1 = rf(ctx, offset, n, details)
	} else {
		r1 = ret.Get(1).(int)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, int, int, bool) error); ok {
		r2 = rf(ctx, offset, n, details)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// UndoChange provides a mock function with given fields: ctx, changeID, userID
func (_m *ExpectationsStore) UndoChange(ctx context.Context, changeID string, userID string) error {
	ret := _m.Called(ctx, changeID, userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, changeID, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
