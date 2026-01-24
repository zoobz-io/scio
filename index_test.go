package scio

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/zoobzio/atom"
	"github.com/zoobzio/grub"
	"github.com/zoobzio/sentinel"
	"github.com/zoobzio/vecna"
)

// mockIndex implements grub.AtomicIndex for testing.
type mockIndex struct {
	spec atom.Spec
	data map[uuid.UUID]*grub.AtomicVector
}

func newMockIndex(spec atom.Spec) *mockIndex {
	return &mockIndex{
		spec: spec,
		data: make(map[uuid.UUID]*grub.AtomicVector),
	}
}

func (m *mockIndex) Spec() atom.Spec { return m.spec }

func (m *mockIndex) Get(_ context.Context, id uuid.UUID) (*grub.AtomicVector, error) {
	if v, ok := m.data[id]; ok {
		return v, nil
	}
	return nil, grub.ErrNotFound
}

func (m *mockIndex) Upsert(_ context.Context, id uuid.UUID, vector []float32, metadata *atom.Atom) error {
	m.data[id] = &grub.AtomicVector{
		ID:       id,
		Vector:   vector,
		Metadata: metadata,
	}
	return nil
}

func (m *mockIndex) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := m.data[id]; !ok {
		return grub.ErrNotFound
	}
	delete(m.data, id)
	return nil
}

func (m *mockIndex) Exists(_ context.Context, id uuid.UUID) (bool, error) {
	_, ok := m.data[id]
	return ok, nil
}

func (m *mockIndex) Search(_ context.Context, _ []float32, k int, _ *atom.Atom) ([]grub.AtomicVector, error) {
	result := make([]grub.AtomicVector, 0, k)
	for _, v := range m.data {
		if len(result) >= k {
			break
		}
		result = append(result, *v)
	}
	return result, nil
}

func (m *mockIndex) Query(_ context.Context, _ []float32, k int, _ *vecna.Filter) ([]grub.AtomicVector, error) {
	result := make([]grub.AtomicVector, 0, k)
	for _, v := range m.data {
		if len(result) >= k {
			break
		}
		result = append(result, *v)
	}
	return result, nil
}

func (m *mockIndex) Filter(_ context.Context, _ *vecna.Filter, limit int) ([]grub.AtomicVector, error) {
	result := make([]grub.AtomicVector, 0, limit)
	for _, v := range m.data {
		if len(result) >= limit {
			break
		}
		result = append(result, *v)
	}
	return result, nil
}

var embeddingSpec = atom.Spec{
	FQDN: "github.com/example/app.Embedding",
	Fields: []sentinel.FieldMetadata{
		{Name: "Text"},
		{Name: "Source"},
	},
}

func TestRegisterIndex(t *testing.T) {
	s := New()
	index := newMockIndex(embeddingSpec)

	err := s.RegisterIndex("idx://embeddings", index, WithDescription("Vector embeddings"))
	if err != nil {
		t.Fatalf("RegisterIndex failed: %v", err)
	}

	resources := s.Indexes()
	if len(resources) != 1 {
		t.Fatalf("expected 1 index, got %d", len(resources))
	}
	if resources[0].Name != "embeddings" {
		t.Errorf("expected name 'embeddings', got %q", resources[0].Name)
	}
	if resources[0].Metadata.Description != "Vector embeddings" {
		t.Errorf("expected description 'Vector embeddings', got %q", resources[0].Metadata.Description)
	}
}

func TestRegisterIndex_DuplicateError(t *testing.T) {
	s := New()
	index := newMockIndex(embeddingSpec)

	_ = s.RegisterIndex("idx://embeddings", index)
	err := s.RegisterIndex("idx://embeddings", index)

	if !errors.Is(err, ErrResourceExists) {
		t.Errorf("expected ErrResourceExists, got %v", err)
	}
}

func TestRegisterIndex_VariantMismatch(t *testing.T) {
	s := New()
	index := newMockIndex(embeddingSpec)

	err := s.RegisterIndex("db://embeddings", index)
	if !errors.Is(err, ErrVariantMismatch) {
		t.Errorf("expected ErrVariantMismatch, got %v", err)
	}
}

func TestGetVector(t *testing.T) {
	s := New()
	index := newMockIndex(embeddingSpec)
	_ = s.RegisterIndex("idx://embeddings", index)

	ctx := context.Background()
	id := uuid.New()
	testVector := []float32{0.1, 0.2, 0.3}
	testMeta := &atom.Atom{Strings: map[string]string{"Text": "hello"}}
	index.data[id] = &grub.AtomicVector{ID: id, Vector: testVector, Metadata: testMeta}

	got, err := s.GetVector(ctx, "idx://embeddings/"+id.String())
	if err != nil {
		t.Fatalf("GetVector failed: %v", err)
	}
	if got.Metadata.Strings["Text"] != "hello" {
		t.Errorf("expected Text=hello, got %v", got.Metadata.Strings["Text"])
	}
}

func TestGetVector_NotFound(t *testing.T) {
	s := New()
	index := newMockIndex(embeddingSpec)
	_ = s.RegisterIndex("idx://embeddings", index)

	ctx := context.Background()
	id := uuid.New()

	_, err := s.GetVector(ctx, "idx://embeddings/"+id.String())
	if !errors.Is(err, grub.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetVector_ResourceNotFound(t *testing.T) {
	s := New()
	ctx := context.Background()
	id := uuid.New()

	_, err := s.GetVector(ctx, "idx://nonexistent/"+id.String())
	if !errors.Is(err, ErrResourceNotFound) {
		t.Errorf("expected ErrResourceNotFound, got %v", err)
	}
}

func TestGetVector_KeyRequired(t *testing.T) {
	s := New()
	index := newMockIndex(embeddingSpec)
	_ = s.RegisterIndex("idx://embeddings", index)

	ctx := context.Background()

	_, err := s.GetVector(ctx, "idx://embeddings")
	if !errors.Is(err, ErrKeyRequired) {
		t.Errorf("expected ErrKeyRequired, got %v", err)
	}
}

func TestGetVector_InvalidUUID(t *testing.T) {
	s := New()
	index := newMockIndex(embeddingSpec)
	_ = s.RegisterIndex("idx://embeddings", index)

	ctx := context.Background()

	_, err := s.GetVector(ctx, "idx://embeddings/not-a-uuid")
	if !errors.Is(err, ErrInvalidUUID) {
		t.Errorf("expected ErrInvalidUUID, got %v", err)
	}
}

func TestGetVector_VariantMismatch(t *testing.T) {
	s := New()
	db := newMockDatabase("users", userSpec)
	_ = s.RegisterDatabase("db://users", db)

	ctx := context.Background()
	id := uuid.New()

	_, err := s.GetVector(ctx, "db://users/"+id.String())
	if !errors.Is(err, ErrVariantMismatch) {
		t.Errorf("expected ErrVariantMismatch, got %v", err)
	}
}

func TestUpsertVector(t *testing.T) {
	s := New()
	index := newMockIndex(embeddingSpec)
	_ = s.RegisterIndex("idx://embeddings", index)

	ctx := context.Background()
	id := uuid.New()
	testVector := []float32{0.1, 0.2, 0.3}
	testMeta := &atom.Atom{Strings: map[string]string{"Text": "world"}}

	err := s.UpsertVector(ctx, "idx://embeddings/"+id.String(), testVector, testMeta)
	if err != nil {
		t.Fatalf("UpsertVector failed: %v", err)
	}

	if index.data[id].Metadata.Strings["Text"] != "world" {
		t.Errorf("expected Text=world in stored data")
	}
}

func TestUpsertVector_VariantMismatch(t *testing.T) {
	s := New()
	store := newMockStore(sessionSpec)
	_ = s.RegisterStore("kv://sessions", store)

	ctx := context.Background()
	id := uuid.New()

	err := s.UpsertVector(ctx, "kv://sessions/"+id.String(), []float32{0.1}, nil)
	if !errors.Is(err, ErrVariantMismatch) {
		t.Errorf("expected ErrVariantMismatch, got %v", err)
	}
}

func TestDeleteVector(t *testing.T) {
	s := New()
	index := newMockIndex(embeddingSpec)
	_ = s.RegisterIndex("idx://embeddings", index)

	ctx := context.Background()
	id := uuid.New()
	index.data[id] = &grub.AtomicVector{ID: id}

	err := s.DeleteVector(ctx, "idx://embeddings/"+id.String())
	if err != nil {
		t.Fatalf("DeleteVector failed: %v", err)
	}

	if _, ok := index.data[id]; ok {
		t.Error("expected data to be deleted")
	}
}

func TestDeleteVector_NotFound(t *testing.T) {
	s := New()
	index := newMockIndex(embeddingSpec)
	_ = s.RegisterIndex("idx://embeddings", index)

	ctx := context.Background()
	id := uuid.New()

	err := s.DeleteVector(ctx, "idx://embeddings/"+id.String())
	if !errors.Is(err, grub.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestVectorExists(t *testing.T) {
	s := New()
	index := newMockIndex(embeddingSpec)
	_ = s.RegisterIndex("idx://embeddings", index)

	ctx := context.Background()
	id := uuid.New()
	index.data[id] = &grub.AtomicVector{ID: id}

	exists, err := s.VectorExists(ctx, "idx://embeddings/"+id.String())
	if err != nil {
		t.Fatalf("VectorExists failed: %v", err)
	}
	if !exists {
		t.Error("expected exists=true")
	}

	otherID := uuid.New()
	exists, err = s.VectorExists(ctx, "idx://embeddings/"+otherID.String())
	if err != nil {
		t.Fatalf("VectorExists failed: %v", err)
	}
	if exists {
		t.Error("expected exists=false")
	}
}

func TestSearchVectors(t *testing.T) {
	s := New()
	index := newMockIndex(embeddingSpec)
	_ = s.RegisterIndex("idx://embeddings", index)

	ctx := context.Background()
	id1, id2 := uuid.New(), uuid.New()
	index.data[id1] = &grub.AtomicVector{ID: id1, Vector: []float32{0.1, 0.2}}
	index.data[id2] = &grub.AtomicVector{ID: id2, Vector: []float32{0.3, 0.4}}

	results, err := s.SearchVectors(ctx, "idx://embeddings", []float32{0.1, 0.2}, 10, nil)
	if err != nil {
		t.Fatalf("SearchVectors failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestSearchVectors_VariantMismatch(t *testing.T) {
	s := New()
	db := newMockDatabase("users", userSpec)
	_ = s.RegisterDatabase("db://users", db)

	ctx := context.Background()

	_, err := s.SearchVectors(ctx, "db://users", []float32{0.1}, 10, nil)
	if !errors.Is(err, ErrVariantMismatch) {
		t.Errorf("expected ErrVariantMismatch, got %v", err)
	}
}

func TestSearchVectors_KeyNotExpected(t *testing.T) {
	s := New()
	index := newMockIndex(embeddingSpec)
	_ = s.RegisterIndex("idx://embeddings", index)

	ctx := context.Background()
	id := uuid.New()

	_, err := s.SearchVectors(ctx, "idx://embeddings/"+id.String(), []float32{0.1}, 10, nil)
	if !errors.Is(err, ErrKeyNotExpected) {
		t.Errorf("expected ErrKeyNotExpected, got %v", err)
	}
}

func TestQueryVectors(t *testing.T) {
	s := New()
	index := newMockIndex(embeddingSpec)
	_ = s.RegisterIndex("idx://embeddings", index)

	ctx := context.Background()
	id := uuid.New()
	index.data[id] = &grub.AtomicVector{ID: id, Vector: []float32{0.1, 0.2}}

	results, err := s.QueryVectors(ctx, "idx://embeddings", []float32{0.1, 0.2}, 10, nil)
	if err != nil {
		t.Fatalf("QueryVectors failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestQueryVectors_VariantMismatch(t *testing.T) {
	s := New()
	store := newMockStore(sessionSpec)
	_ = s.RegisterStore("kv://sessions", store)

	ctx := context.Background()

	_, err := s.QueryVectors(ctx, "kv://sessions", []float32{0.1}, 10, nil)
	if !errors.Is(err, ErrVariantMismatch) {
		t.Errorf("expected ErrVariantMismatch, got %v", err)
	}
}

func TestFilterVectors(t *testing.T) {
	s := New()
	index := newMockIndex(embeddingSpec)
	_ = s.RegisterIndex("idx://embeddings", index)

	ctx := context.Background()
	id := uuid.New()
	index.data[id] = &grub.AtomicVector{ID: id, Vector: []float32{0.1, 0.2}}

	results, err := s.FilterVectors(ctx, "idx://embeddings", nil, 10)
	if err != nil {
		t.Fatalf("FilterVectors failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestFilterVectors_VariantMismatch(t *testing.T) {
	s := New()
	bucket := newMockBucket(documentSpec)
	_ = s.RegisterBucket("bcs://documents", bucket)

	ctx := context.Background()

	_, err := s.FilterVectors(ctx, "bcs://documents", nil, 10)
	if !errors.Is(err, ErrVariantMismatch) {
		t.Errorf("expected ErrVariantMismatch, got %v", err)
	}
}

func TestIndexes_InSources(t *testing.T) {
	s := New()
	_ = s.RegisterDatabase("db://users", newMockDatabase("users", userSpec))
	_ = s.RegisterIndex("idx://embeddings", newMockIndex(embeddingSpec))

	sources := s.Sources()
	if len(sources) != 2 {
		t.Errorf("expected 2 sources, got %d", len(sources))
	}
}
