// Package testing provides test utilities for scio.
package testing

import (
	"context"
	"testing"
	"time"

	"github.com/zoobzio/atom"
	"github.com/zoobzio/edamame"
	"github.com/zoobzio/grub"
	"github.com/zoobzio/sentinel"
)

// WithTimeout creates a context with timeout for tests.
func WithTimeout(t *testing.T, d time.Duration) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), d)
	t.Cleanup(cancel)
	return ctx
}

// AssertNoError fails the test if err is not nil.
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AssertError fails the test if err is nil.
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// AssertEqual fails the test if got != want.
func AssertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

// AssertTrue fails the test if v is false.
func AssertTrue(t *testing.T, v bool) {
	t.Helper()
	if !v {
		t.Error("expected true, got false")
	}
}

// AssertFalse fails the test if v is true.
func AssertFalse(t *testing.T, v bool) {
	t.Helper()
	if v {
		t.Error("expected false, got true")
	}
}

// MockDatabase implements grub.AtomicDatabase for testing.
type MockDatabase struct {
	TableName string
	TypeSpec  atom.Spec
	Data      map[string]*atom.Atom
}

// NewMockDatabase creates a new mock database.
func NewMockDatabase(table string, spec atom.Spec) *MockDatabase {
	return &MockDatabase{
		TableName: table,
		TypeSpec:  spec,
		Data:      make(map[string]*atom.Atom),
	}
}

// Table returns the table name.
func (m *MockDatabase) Table() string { return m.TableName }

// Spec returns the atom spec.
func (m *MockDatabase) Spec() atom.Spec { return m.TypeSpec }

// Get retrieves a record by key.
func (m *MockDatabase) Get(_ context.Context, key string) (*atom.Atom, error) {
	if a, ok := m.Data[key]; ok {
		return a, nil
	}
	return nil, grub.ErrNotFound
}

// Set stores a record at key.
func (m *MockDatabase) Set(_ context.Context, key string, data *atom.Atom) error {
	m.Data[key] = data
	return nil
}

// Delete removes a record at key.
func (m *MockDatabase) Delete(_ context.Context, key string) error {
	if _, ok := m.Data[key]; !ok {
		return grub.ErrNotFound
	}
	delete(m.Data, key)
	return nil
}

// Exists checks if a record exists at key.
func (m *MockDatabase) Exists(_ context.Context, key string) (bool, error) {
	_, ok := m.Data[key]
	return ok, nil
}

// Query returns all records.
func (m *MockDatabase) Query(_ context.Context, _ edamame.QueryStatement, _ map[string]any) ([]*atom.Atom, error) {
	var result []*atom.Atom
	for _, a := range m.Data {
		result = append(result, a)
	}
	return result, nil
}

// Select returns the first record.
func (m *MockDatabase) Select(_ context.Context, _ edamame.SelectStatement, _ map[string]any) (*atom.Atom, error) {
	for _, a := range m.Data {
		return a, nil
	}
	return nil, grub.ErrNotFound
}

// MockStore implements grub.AtomicStore for testing.
type MockStore struct {
	TypeSpec atom.Spec
	Data     map[string]*atom.Atom
}

// NewMockStore creates a new mock store.
func NewMockStore(spec atom.Spec) *MockStore {
	return &MockStore{
		TypeSpec: spec,
		Data:     make(map[string]*atom.Atom),
	}
}

// Spec returns the atom spec.
func (m *MockStore) Spec() atom.Spec { return m.TypeSpec }

// Get retrieves a value by key.
func (m *MockStore) Get(_ context.Context, key string) (*atom.Atom, error) {
	if a, ok := m.Data[key]; ok {
		return a, nil
	}
	return nil, grub.ErrNotFound
}

// Set stores a value at key.
func (m *MockStore) Set(_ context.Context, key string, data *atom.Atom, _ time.Duration) error {
	m.Data[key] = data
	return nil
}

// Delete removes a value at key.
func (m *MockStore) Delete(_ context.Context, key string) error {
	if _, ok := m.Data[key]; !ok {
		return grub.ErrNotFound
	}
	delete(m.Data, key)
	return nil
}

// Exists checks if a value exists at key.
func (m *MockStore) Exists(_ context.Context, key string) (bool, error) {
	_, ok := m.Data[key]
	return ok, nil
}

// MockBucket implements grub.AtomicBucket for testing.
type MockBucket struct {
	TypeSpec atom.Spec
	Data     map[string]*grub.AtomicObject
}

// NewMockBucket creates a new mock bucket.
func NewMockBucket(spec atom.Spec) *MockBucket {
	return &MockBucket{
		TypeSpec: spec,
		Data:     make(map[string]*grub.AtomicObject),
	}
}

// Spec returns the atom spec.
func (m *MockBucket) Spec() atom.Spec { return m.TypeSpec }

// Get retrieves a blob by key.
func (m *MockBucket) Get(_ context.Context, key string) (*grub.AtomicObject, error) {
	if obj, ok := m.Data[key]; ok {
		return obj, nil
	}
	return nil, grub.ErrNotFound
}

// Put stores a blob at key.
func (m *MockBucket) Put(_ context.Context, key string, obj *grub.AtomicObject) error {
	m.Data[key] = obj
	return nil
}

// Delete removes a blob at key.
func (m *MockBucket) Delete(_ context.Context, key string) error {
	if _, ok := m.Data[key]; !ok {
		return grub.ErrNotFound
	}
	delete(m.Data, key)
	return nil
}

// Exists checks if a blob exists at key.
func (m *MockBucket) Exists(_ context.Context, key string) (bool, error) {
	_, ok := m.Data[key]
	return ok, nil
}

// TestSpec creates a test spec with the given name and fields.
func TestSpec(name string, fields ...string) atom.Spec {
	fieldMeta := make([]sentinel.FieldMetadata, len(fields))
	for i, f := range fields {
		fieldMeta[i] = sentinel.FieldMetadata{Name: f}
	}
	return atom.Spec{
		FQDN:   "github.com/test." + name,
		Fields: fieldMeta,
	}
}
