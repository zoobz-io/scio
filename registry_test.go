package scio

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zoobzio/atom"
	"github.com/zoobzio/edamame"
	"github.com/zoobzio/grub"
	"github.com/zoobzio/sentinel"
)

// mockDatabase implements grub.AtomicDatabase for testing.
type mockDatabase struct {
	table string
	spec  atom.Spec
	data  map[string]*atom.Atom
}

func newMockDatabase(table string, spec atom.Spec) *mockDatabase {
	return &mockDatabase{
		table: table,
		spec:  spec,
		data:  make(map[string]*atom.Atom),
	}
}

func (m *mockDatabase) Table() string               { return m.table }
func (m *mockDatabase) Spec() atom.Spec             { return m.spec }
func (m *mockDatabase) Get(_ context.Context, key string) (*atom.Atom, error) {
	if a, ok := m.data[key]; ok {
		return a, nil
	}
	return nil, grub.ErrNotFound
}
func (m *mockDatabase) Set(_ context.Context, key string, data *atom.Atom) error {
	m.data[key] = data
	return nil
}
func (m *mockDatabase) Delete(_ context.Context, key string) error {
	if _, ok := m.data[key]; !ok {
		return grub.ErrNotFound
	}
	delete(m.data, key)
	return nil
}
func (m *mockDatabase) Exists(_ context.Context, key string) (bool, error) {
	_, ok := m.data[key]
	return ok, nil
}
func (m *mockDatabase) ExecQuery(_ context.Context, _ edamame.QueryStatement, _ map[string]any) ([]*atom.Atom, error) {
	result := make([]*atom.Atom, 0, len(m.data))
	for _, a := range m.data {
		result = append(result, a)
	}
	return result, nil
}
func (m *mockDatabase) ExecSelect(_ context.Context, _ edamame.SelectStatement, _ map[string]any) (*atom.Atom, error) {
	for _, a := range m.data {
		return a, nil
	}
	return nil, grub.ErrNotFound
}

// mockStore implements grub.AtomicStore for testing.
type mockStore struct {
	spec atom.Spec
	data map[string]*atom.Atom
}

func newMockStore(spec atom.Spec) *mockStore {
	return &mockStore{
		spec: spec,
		data: make(map[string]*atom.Atom),
	}
}

func (m *mockStore) Spec() atom.Spec { return m.spec }
func (m *mockStore) Get(_ context.Context, key string) (*atom.Atom, error) {
	if a, ok := m.data[key]; ok {
		return a, nil
	}
	return nil, grub.ErrNotFound
}
func (m *mockStore) Set(_ context.Context, key string, data *atom.Atom, _ time.Duration) error {
	m.data[key] = data
	return nil
}
func (m *mockStore) Delete(_ context.Context, key string) error {
	if _, ok := m.data[key]; !ok {
		return grub.ErrNotFound
	}
	delete(m.data, key)
	return nil
}
func (m *mockStore) Exists(_ context.Context, key string) (bool, error) {
	_, ok := m.data[key]
	return ok, nil
}

// mockBucket implements grub.AtomicBucket for testing.
type mockBucket struct {
	spec atom.Spec
	data map[string]*grub.AtomicObject
}

func newMockBucket(spec atom.Spec) *mockBucket {
	return &mockBucket{
		spec: spec,
		data: make(map[string]*grub.AtomicObject),
	}
}

func (m *mockBucket) Spec() atom.Spec { return m.spec }
func (m *mockBucket) Get(_ context.Context, key string) (*grub.AtomicObject, error) {
	if obj, ok := m.data[key]; ok {
		return obj, nil
	}
	return nil, grub.ErrNotFound
}
func (m *mockBucket) Put(_ context.Context, key string, obj *grub.AtomicObject) error {
	m.data[key] = obj
	return nil
}
func (m *mockBucket) Delete(_ context.Context, key string) error {
	if _, ok := m.data[key]; !ok {
		return grub.ErrNotFound
	}
	delete(m.data, key)
	return nil
}
func (m *mockBucket) Exists(_ context.Context, key string) (bool, error) {
	_, ok := m.data[key]
	return ok, nil
}

// Test specs
var (
	userSpec = atom.Spec{
		FQDN: "github.com/example/app.User",
		Fields: []sentinel.FieldMetadata{
			{Name: "ID"},
			{Name: "Email"},
			{Name: "Name"},
		},
	}
	sessionSpec = atom.Spec{
		FQDN: "github.com/example/app.Session",
		Fields: []sentinel.FieldMetadata{
			{Name: "ID"},
			{Name: "UserID"},
			{Name: "Token"},
		},
	}
	documentSpec = atom.Spec{
		FQDN: "github.com/example/app.Document",
		Fields: []sentinel.FieldMetadata{
			{Name: "Title"},
			{Name: "Content"},
		},
	}
)

func TestRegisterDatabase(t *testing.T) {
	s := New()
	db := newMockDatabase("users", userSpec)

	err := s.RegisterDatabase("db://users", db, WithDescription("User table"))
	if err != nil {
		t.Fatalf("RegisterDatabase failed: %v", err)
	}

	// Verify registration
	resources := s.Databases()
	if len(resources) != 1 {
		t.Fatalf("expected 1 database, got %d", len(resources))
	}
	if resources[0].Name != "users" {
		t.Errorf("expected name 'users', got %q", resources[0].Name)
	}
	if resources[0].Metadata.Description != "User table" {
		t.Errorf("expected description 'User table', got %q", resources[0].Metadata.Description)
	}
}

func TestRegisterDatabase_DuplicateError(t *testing.T) {
	s := New()
	db := newMockDatabase("users", userSpec)

	_ = s.RegisterDatabase("db://users", db)
	err := s.RegisterDatabase("db://users", db)

	if !errors.Is(err, ErrResourceExists) {
		t.Errorf("expected ErrResourceExists, got %v", err)
	}
}

func TestRegisterDatabase_VariantMismatch(t *testing.T) {
	s := New()
	db := newMockDatabase("users", userSpec)

	err := s.RegisterDatabase("kv://users", db)
	if !errors.Is(err, ErrVariantMismatch) {
		t.Errorf("expected ErrVariantMismatch, got %v", err)
	}
}

func TestRegisterStore(t *testing.T) {
	s := New()
	store := newMockStore(sessionSpec)

	err := s.RegisterStore("kv://sessions", store, WithTag("type", "cache"))
	if err != nil {
		t.Fatalf("RegisterStore failed: %v", err)
	}

	resources := s.Stores()
	if len(resources) != 1 {
		t.Fatalf("expected 1 store, got %d", len(resources))
	}
	if resources[0].Metadata.Tags["type"] != "cache" {
		t.Errorf("expected tag type=cache, got %v", resources[0].Metadata.Tags)
	}
}

func TestRegisterBucket(t *testing.T) {
	s := New()
	bucket := newMockBucket(documentSpec)

	err := s.RegisterBucket("bcs://documents", bucket, WithVersion("1.0"))
	if err != nil {
		t.Fatalf("RegisterBucket failed: %v", err)
	}

	resources := s.Buckets()
	if len(resources) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(resources))
	}
	if resources[0].Metadata.Version != "1.0" {
		t.Errorf("expected version '1.0', got %q", resources[0].Metadata.Version)
	}
}

func TestSources(t *testing.T) {
	s := New()
	_ = s.RegisterDatabase("db://users", newMockDatabase("users", userSpec))
	_ = s.RegisterStore("kv://sessions", newMockStore(sessionSpec))
	_ = s.RegisterBucket("bcs://documents", newMockBucket(documentSpec))

	sources := s.Sources()
	if len(sources) != 3 {
		t.Errorf("expected 3 sources, got %d", len(sources))
	}
}

func TestFindBySpec(t *testing.T) {
	s := New()

	// Register two resources with the same spec
	_ = s.RegisterDatabase("db://users", newMockDatabase("users", userSpec))
	_ = s.RegisterStore("kv://user-cache", newMockStore(userSpec))

	// Register one with different spec
	_ = s.RegisterStore("kv://sessions", newMockStore(sessionSpec))

	found := s.FindBySpec(userSpec)
	if len(found) != 2 {
		t.Errorf("expected 2 resources with userSpec, got %d", len(found))
	}
}

func TestFindByField(t *testing.T) {
	s := New()
	_ = s.RegisterDatabase("db://users", newMockDatabase("users", userSpec))
	_ = s.RegisterStore("kv://sessions", newMockStore(sessionSpec))

	// Both specs have "ID" field
	found := s.FindByField("ID")
	if len(found) != 2 {
		t.Errorf("expected 2 resources with ID field, got %d", len(found))
	}

	// Only userSpec has "Email" field
	found = s.FindByField("Email")
	if len(found) != 1 {
		t.Errorf("expected 1 resource with Email field, got %d", len(found))
	}

	// No spec has "Nonexistent" field
	found = s.FindByField("Nonexistent")
	if len(found) != 0 {
		t.Errorf("expected 0 resources with Nonexistent field, got %d", len(found))
	}
}

func TestRelated(t *testing.T) {
	s := New()

	// Register two resources with the same spec
	_ = s.RegisterDatabase("db://users", newMockDatabase("users", userSpec))
	_ = s.RegisterStore("kv://user-cache", newMockStore(userSpec))

	// Related should return the other resource, not itself
	related := s.Related("db://users")
	if len(related) != 1 {
		t.Fatalf("expected 1 related resource, got %d", len(related))
	}
	if related[0].URI != "kv://user-cache" {
		t.Errorf("expected related URI 'kv://user-cache', got %q", related[0].URI)
	}
}

func TestSpec(t *testing.T) {
	s := New()
	_ = s.RegisterDatabase("db://users", newMockDatabase("users", userSpec))

	spec, err := s.Spec("db://users")
	if err != nil {
		t.Fatalf("Spec failed: %v", err)
	}
	if spec.FQDN != userSpec.FQDN {
		t.Errorf("expected FQDN %q, got %q", userSpec.FQDN, spec.FQDN)
	}
}

func TestSpec_NotFound(t *testing.T) {
	s := New()

	_, err := s.Spec("db://nonexistent")
	if !errors.Is(err, ErrResourceNotFound) {
		t.Errorf("expected ErrResourceNotFound, got %v", err)
	}
}

func TestResource(t *testing.T) {
	s := New()
	_ = s.RegisterDatabase("db://users", newMockDatabase("users", userSpec), WithDescription("test"))

	r := s.Resource("db://users")
	if r == nil {
		t.Fatal("expected resource, got nil")
	}
	if r.Name != "users" {
		t.Errorf("expected name 'users', got %q", r.Name)
	}
	if r.Metadata.Description != "test" {
		t.Errorf("expected description 'test', got %q", r.Metadata.Description)
	}
}

func TestResource_NotFound(t *testing.T) {
	s := New()

	r := s.Resource("db://nonexistent")
	if r != nil {
		t.Errorf("expected nil for nonexistent resource, got %v", r)
	}
}
