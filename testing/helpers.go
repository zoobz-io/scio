// Package testing provides test utilities for scio.
package testing

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zoobz-io/atom"
	"github.com/zoobz-io/edamame"
	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/lucene"
	"github.com/zoobz-io/sentinel"
	"github.com/zoobz-io/vecna"
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

// MockIndex implements grub.AtomicIndex for testing.
type MockIndex struct {
	TypeSpec atom.Spec
	Data     map[uuid.UUID]*grub.AtomicVector
}

// NewMockIndex creates a new mock index.
func NewMockIndex(spec atom.Spec) *MockIndex {
	return &MockIndex{
		TypeSpec: spec,
		Data:     make(map[uuid.UUID]*grub.AtomicVector),
	}
}

// Spec returns the atom spec.
func (m *MockIndex) Spec() atom.Spec { return m.TypeSpec }

// Get retrieves a vector by ID.
func (m *MockIndex) Get(_ context.Context, id uuid.UUID) (*grub.AtomicVector, error) {
	if v, ok := m.Data[id]; ok {
		return v, nil
	}
	return nil, grub.ErrNotFound
}

// Upsert stores a vector at ID.
func (m *MockIndex) Upsert(_ context.Context, id uuid.UUID, vector []float32, metadata *atom.Atom) error {
	m.Data[id] = &grub.AtomicVector{
		ID:       id,
		Vector:   vector,
		Metadata: metadata,
	}
	return nil
}

// Delete removes a vector at ID.
func (m *MockIndex) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := m.Data[id]; !ok {
		return grub.ErrNotFound
	}
	delete(m.Data, id)
	return nil
}

// Exists checks if a vector exists at ID.
func (m *MockIndex) Exists(_ context.Context, id uuid.UUID) (bool, error) {
	_, ok := m.Data[id]
	return ok, nil
}

// Search performs similarity search returning atomized results.
func (m *MockIndex) Search(_ context.Context, _ []float32, k int, _ *atom.Atom) ([]grub.AtomicVector, error) {
	var result []grub.AtomicVector
	for _, v := range m.Data {
		if len(result) >= k {
			break
		}
		result = append(result, *v)
	}
	return result, nil
}

// Query performs similarity search with vecna filter support.
func (m *MockIndex) Query(_ context.Context, _ []float32, k int, _ *vecna.Filter) ([]grub.AtomicVector, error) {
	var result []grub.AtomicVector
	for _, v := range m.Data {
		if len(result) >= k {
			break
		}
		result = append(result, *v)
	}
	return result, nil
}

// Filter returns vectors matching the metadata filter without similarity search.
func (m *MockIndex) Filter(_ context.Context, _ *vecna.Filter, limit int) ([]grub.AtomicVector, error) {
	var result []grub.AtomicVector
	for _, v := range m.Data {
		if len(result) >= limit {
			break
		}
		result = append(result, *v)
	}
	return result, nil
}

// MockSearch implements grub.AtomicSearch for testing.
type MockSearch struct {
	IndexName string
	TypeSpec  atom.Spec
	Data      map[string]*grub.AtomicDocument
}

// NewMockSearch creates a new mock search.
func NewMockSearch(indexName string, spec atom.Spec) *MockSearch {
	return &MockSearch{
		IndexName: indexName,
		TypeSpec:  spec,
		Data:      make(map[string]*grub.AtomicDocument),
	}
}

// Index returns the index name.
func (m *MockSearch) Index() string { return m.IndexName }

// Spec returns the atom spec.
func (m *MockSearch) Spec() atom.Spec { return m.TypeSpec }

// Get retrieves a document by ID.
func (m *MockSearch) Get(_ context.Context, id string) (*grub.AtomicDocument, error) {
	if doc, ok := m.Data[id]; ok {
		return doc, nil
	}
	return nil, grub.ErrNotFound
}

// IndexDoc stores a document at ID.
func (m *MockSearch) IndexDoc(_ context.Context, id string, doc *atom.Atom) error {
	m.Data[id] = &grub.AtomicDocument{
		ID:      id,
		Content: doc,
	}
	return nil
}

// Delete removes a document at ID.
func (m *MockSearch) Delete(_ context.Context, id string) error {
	if _, ok := m.Data[id]; !ok {
		return grub.ErrNotFound
	}
	delete(m.Data, id)
	return nil
}

// Exists checks if a document exists at ID.
func (m *MockSearch) Exists(_ context.Context, id string) (bool, error) {
	_, ok := m.Data[id]
	return ok, nil
}

// Search performs a search returning all documents.
func (m *MockSearch) Search(_ context.Context, _ *lucene.Search) ([]grub.AtomicDocument, error) {
	result := make([]grub.AtomicDocument, 0, len(m.Data))
	for _, doc := range m.Data {
		result = append(result, *doc)
	}
	return result, nil
}
