// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	expectations "go.skia.org/infra/golden/go/expectations"

	types "go.skia.org/infra/golden/go/types"
)

// Store is an autogenerated mock type for the Store type
type Store struct {
	mock.Mock
}

// AddChange provides a mock function with given fields: ctx, changes, userId
func (_m *Store) AddChange(ctx context.Context, changes []expectations.Delta, userId string) error {
	ret := _m.Called(ctx, changes, userId)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []expectations.Delta, string) error); ok {
		r0 = rf(ctx, changes, userId)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ForChangeList provides a mock function with given fields: id, crs
func (_m *Store) ForChangeList(id string, crs string) expectations.Store {
	ret := _m.Called(id, crs)

	var r0 expectations.Store
	if rf, ok := ret.Get(0).(func(string, string) expectations.Store); ok {
		r0 = rf(id, crs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(expectations.Store)
		}
	}

	return r0
}

// Get provides a mock function with given fields: ctx
func (_m *Store) Get(ctx context.Context) (expectations.ReadOnly, error) {
	ret := _m.Called(ctx)

	var r0 expectations.ReadOnly
	if rf, ok := ret.Get(0).(func(context.Context) expectations.ReadOnly); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(expectations.ReadOnly)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCopy provides a mock function with given fields: ctx
func (_m *Store) GetCopy(ctx context.Context) (*expectations.Expectations, error) {
	ret := _m.Called(ctx)

	var r0 *expectations.Expectations
	if rf, ok := ret.Get(0).(func(context.Context) *expectations.Expectations); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*expectations.Expectations)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTriageHistory provides a mock function with given fields: ctx, grouping, digest
func (_m *Store) GetTriageHistory(ctx context.Context, grouping types.TestName, digest types.Digest) ([]expectations.TriageHistory, error) {
	ret := _m.Called(ctx, grouping, digest)

	var r0 []expectations.TriageHistory
	if rf, ok := ret.Get(0).(func(context.Context, types.TestName, types.Digest) []expectations.TriageHistory); ok {
		r0 = rf(ctx, grouping, digest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]expectations.TriageHistory)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, types.TestName, types.Digest) error); ok {
		r1 = rf(ctx, grouping, digest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// QueryLog provides a mock function with given fields: ctx, offset, n, details
func (_m *Store) QueryLog(ctx context.Context, offset int, n int, details bool) ([]expectations.TriageLogEntry, int, error) {
	ret := _m.Called(ctx, offset, n, details)

	var r0 []expectations.TriageLogEntry
	if rf, ok := ret.Get(0).(func(context.Context, int, int, bool) []expectations.TriageLogEntry); ok {
		r0 = rf(ctx, offset, n, details)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]expectations.TriageLogEntry)
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
func (_m *Store) UndoChange(ctx context.Context, changeID string, userID string) error {
	ret := _m.Called(ctx, changeID, userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, changeID, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
