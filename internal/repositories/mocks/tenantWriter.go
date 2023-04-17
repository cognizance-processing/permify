package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	base "permify/pkg/pb/base/v1"
)

// TenantWriter is an autogenerated mock type for the TenantWriter type
type TenantWriter struct {
	mock.Mock
}

// CreateTenant - Create Tenant to repository
func (_m *TenantWriter) CreateTenant(ctx context.Context, id, name string) (tenant *base.Tenant, err error) {
	ret := _m.Called(name)

	var r0 *base.Tenant
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *base.Tenant); ok {
		r0 = rf(ctx, id, name)
	} else {
		r0 = ret.Get(0).(*base.Tenant)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, id, name)
	} else {
		if e, ok := ret.Get(1).(error); ok {
			r1 = e
		} else {
			r1 = nil
		}
	}

	return r0, r1
}

// DeleteTenant - Delete Tenant to repository
func (_m *TenantWriter) DeleteTenant(ctx context.Context, tenantID string) (tenant *base.Tenant, err error) {
	ret := _m.Called(tenantID)

	var r0 *base.Tenant
	if rf, ok := ret.Get(0).(func(context.Context, string) *base.Tenant); ok {
		r0 = rf(ctx, tenantID)
	} else {
		r0 = ret.Get(0).(*base.Tenant)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, tenantID)
	} else {
		if e, ok := ret.Get(1).(error); ok {
			r1 = e
		} else {
			r1 = nil
		}
	}

	return r0, r1
}
