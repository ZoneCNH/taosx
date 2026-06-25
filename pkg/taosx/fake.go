package taosx

import (
	"context"
	"sync"
	"time"
)

type FakeDriver struct {
	ExecResult      ExecResult
	QueryRows       Rows
	WriteResult     WriteResult
	DeleteResult    ExecResult
	ExecError       error
	QueryError      error
	WriteError      error
	DeleteError     error
	SchemalessError error
	HealthError     error
	CloseError      error

	mu              sync.Mutex
	execCalls       int
	queryCalls      int
	writeCalls      int
	deleteCalls     int
	schemalessCalls int
	healthCalls     int
	closeCalls      int
	closed          bool
}

type FakeClient struct {
	ExecResult      ExecResult
	QueryRows       Rows
	WriteResult     WriteResult
	DeleteResult    ExecResult
	HealthStatus    HealthStatus
	ExecError       error
	QueryError      error
	WriteError      error
	DeleteError     error
	SchemalessError error
	CloseError      error

	mu              sync.Mutex
	execCalls       int
	queryCalls      int
	writeCalls      int
	deleteCalls     int
	schemalessCalls int
	healthCalls     int
	closeCalls      int
	closed          bool
}

func NewFakeDriver() *FakeDriver {
	return &FakeDriver{}
}

func NewFakeClient() *FakeClient {
	return &FakeClient{}
}

func (f *FakeClient) Exec(context.Context, Statement) (ExecResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.execCalls++
	return f.ExecResult, f.ExecError
}

func (f *FakeClient) Query(context.Context, Query) (Rows, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.queryCalls++
	return f.QueryRows, f.QueryError
}

func (f *FakeClient) WriteBatch(context.Context, Batch) (WriteResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.writeCalls++
	return f.WriteResult, f.WriteError
}

func (f *FakeClient) SchemalessWrite(context.Context, SchemalessPayload) (WriteResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.schemalessCalls++
	return f.WriteResult, f.SchemalessError
}

func (f *FakeClient) DeleteRange(context.Context, string, time.Time) (ExecResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deleteCalls++
	return f.DeleteResult, f.DeleteError
}

func (f *FakeClient) Health(context.Context) HealthStatus {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.healthCalls++
	if f.HealthStatus.Status == "" {
		return HealthStatus{Status: HealthHealthy}
	}
	return f.HealthStatus
}

func (f *FakeClient) Close(context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closeCalls++
	f.closed = true
	return f.CloseError
}

func (f *FakeClient) ExecCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.execCalls
}

func (f *FakeClient) QueryCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.queryCalls
}

func (f *FakeClient) WriteCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.writeCalls
}

func (f *FakeClient) SchemalessCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.schemalessCalls
}

func (f *FakeClient) DeleteCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.deleteCalls
}

func (f *FakeClient) HealthCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.healthCalls
}

func (f *FakeClient) CloseCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.closeCalls
}

func (f *FakeClient) Closed() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.closed
}

func (f *FakeDriver) Exec(context.Context, Statement) (ExecResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.execCalls++
	return f.ExecResult, f.ExecError
}

func (f *FakeDriver) Query(context.Context, Query) (Rows, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.queryCalls++
	return f.QueryRows, f.QueryError
}

func (f *FakeDriver) WriteBatch(context.Context, Batch) (WriteResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.writeCalls++
	return f.WriteResult, f.WriteError
}

func (f *FakeDriver) SchemalessWrite(context.Context, SchemalessPayload) (WriteResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.schemalessCalls++
	return f.WriteResult, f.SchemalessError
}

func (f *FakeDriver) DeleteRange(context.Context, string, time.Time) (ExecResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deleteCalls++
	return f.DeleteResult, f.DeleteError
}

func (f *FakeDriver) Health(context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.healthCalls++
	return f.HealthError
}

func (f *FakeDriver) Close(context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closeCalls++
	f.closed = true
	return f.CloseError
}

func (f *FakeDriver) ExecCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.execCalls
}

func (f *FakeDriver) QueryCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.queryCalls
}

func (f *FakeDriver) WriteCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.writeCalls
}

func (f *FakeDriver) SchemalessCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.schemalessCalls
}

func (f *FakeDriver) DeleteCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.deleteCalls
}

func (f *FakeDriver) HealthCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.healthCalls
}

func (f *FakeDriver) CloseCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.closeCalls
}

func (f *FakeDriver) Closed() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.closed
}
