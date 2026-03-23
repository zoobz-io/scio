package scio

import (
	"context"
	"errors"
	"testing"

	"github.com/zoobz-io/atom"
	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/lucene"
	"github.com/zoobz-io/sentinel"
)

// mockSearch implements grub.AtomicSearch for testing.
type mockSearch struct {
	indexName string
	spec      atom.Spec
	data      map[string]*grub.AtomicDocument
}

func newMockSearch(indexName string, spec atom.Spec) *mockSearch {
	return &mockSearch{
		indexName: indexName,
		spec:      spec,
		data:      make(map[string]*grub.AtomicDocument),
	}
}

func (m *mockSearch) Index() string        { return m.indexName }
func (m *mockSearch) Spec() atom.Spec      { return m.spec }

func (m *mockSearch) Get(_ context.Context, id string) (*grub.AtomicDocument, error) {
	if doc, ok := m.data[id]; ok {
		return doc, nil
	}
	return nil, grub.ErrNotFound
}

func (m *mockSearch) IndexDoc(_ context.Context, id string, doc *atom.Atom) error {
	m.data[id] = &grub.AtomicDocument{
		ID:      id,
		Content: doc,
	}
	return nil
}

func (m *mockSearch) Delete(_ context.Context, id string) error {
	if _, ok := m.data[id]; !ok {
		return grub.ErrNotFound
	}
	delete(m.data, id)
	return nil
}

func (m *mockSearch) Exists(_ context.Context, id string) (bool, error) {
	_, ok := m.data[id]
	return ok, nil
}

func (m *mockSearch) Search(_ context.Context, _ *lucene.Search) ([]grub.AtomicDocument, error) {
	result := make([]grub.AtomicDocument, 0, len(m.data))
	for _, doc := range m.data {
		result = append(result, *doc)
	}
	return result, nil
}

var articleSpec = atom.Spec{
	FQDN: "github.com/example/app.Article",
	Fields: []sentinel.FieldMetadata{
		{Name: "Title"},
		{Name: "Body"},
		{Name: "Author"},
	},
}

func TestRegisterSearch(t *testing.T) {
	s := New()
	search := newMockSearch("articles", articleSpec)

	err := s.RegisterSearch("srch://articles", search, WithDescription("Article search"))
	if err != nil {
		t.Fatalf("RegisterSearch failed: %v", err)
	}

	resources := s.Searches()
	if len(resources) != 1 {
		t.Fatalf("expected 1 search, got %d", len(resources))
	}
	if resources[0].Name != "articles" {
		t.Errorf("expected name 'articles', got %q", resources[0].Name)
	}
	if resources[0].Metadata.Description != "Article search" {
		t.Errorf("expected description 'Article search', got %q", resources[0].Metadata.Description)
	}
}

func TestRegisterSearch_DuplicateError(t *testing.T) {
	s := New()
	search := newMockSearch("articles", articleSpec)

	_ = s.RegisterSearch("srch://articles", search)
	err := s.RegisterSearch("srch://articles", search)

	if !errors.Is(err, ErrResourceExists) {
		t.Errorf("expected ErrResourceExists, got %v", err)
	}
}

func TestRegisterSearch_VariantMismatch(t *testing.T) {
	s := New()
	search := newMockSearch("articles", articleSpec)

	err := s.RegisterSearch("db://articles", search)
	if !errors.Is(err, ErrVariantMismatch) {
		t.Errorf("expected ErrVariantMismatch, got %v", err)
	}
}

func TestGetDocument(t *testing.T) {
	s := New()
	search := newMockSearch("articles", articleSpec)
	_ = s.RegisterSearch("srch://articles", search)

	ctx := context.Background()
	testContent := &atom.Atom{Strings: map[string]string{"Title": "hello"}}
	search.data["doc-1"] = &grub.AtomicDocument{ID: "doc-1", Content: testContent}

	got, err := s.GetDocument(ctx, "srch://articles/doc-1")
	if err != nil {
		t.Fatalf("GetDocument failed: %v", err)
	}
	if got.Content.Strings["Title"] != "hello" {
		t.Errorf("expected Title=hello, got %v", got.Content.Strings["Title"])
	}
}

func TestGetDocument_NotFound(t *testing.T) {
	s := New()
	search := newMockSearch("articles", articleSpec)
	_ = s.RegisterSearch("srch://articles", search)

	ctx := context.Background()

	_, err := s.GetDocument(ctx, "srch://articles/nonexistent")
	if !errors.Is(err, grub.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetDocument_ResourceNotFound(t *testing.T) {
	s := New()
	ctx := context.Background()

	_, err := s.GetDocument(ctx, "srch://nonexistent/doc-1")
	if !errors.Is(err, ErrResourceNotFound) {
		t.Errorf("expected ErrResourceNotFound, got %v", err)
	}
}

func TestGetDocument_KeyRequired(t *testing.T) {
	s := New()
	search := newMockSearch("articles", articleSpec)
	_ = s.RegisterSearch("srch://articles", search)

	ctx := context.Background()

	_, err := s.GetDocument(ctx, "srch://articles")
	if !errors.Is(err, ErrKeyRequired) {
		t.Errorf("expected ErrKeyRequired, got %v", err)
	}
}

func TestGetDocument_VariantMismatch(t *testing.T) {
	s := New()
	db := newMockDatabase("users", userSpec)
	_ = s.RegisterDatabase("db://users", db)

	ctx := context.Background()

	_, err := s.GetDocument(ctx, "db://users/123")
	if !errors.Is(err, ErrVariantMismatch) {
		t.Errorf("expected ErrVariantMismatch, got %v", err)
	}
}

func TestIndexDocument(t *testing.T) {
	s := New()
	search := newMockSearch("articles", articleSpec)
	_ = s.RegisterSearch("srch://articles", search)

	ctx := context.Background()
	testContent := &atom.Atom{Strings: map[string]string{"Title": "world"}}

	err := s.IndexDocument(ctx, "srch://articles/doc-1", testContent)
	if err != nil {
		t.Fatalf("IndexDocument failed: %v", err)
	}

	if search.data["doc-1"].Content.Strings["Title"] != "world" {
		t.Errorf("expected Title=world in stored data")
	}
}

func TestIndexDocument_VariantMismatch(t *testing.T) {
	s := New()
	store := newMockStore(sessionSpec)
	_ = s.RegisterStore("kv://sessions", store)

	ctx := context.Background()

	err := s.IndexDocument(ctx, "kv://sessions/key1", nil)
	if !errors.Is(err, ErrVariantMismatch) {
		t.Errorf("expected ErrVariantMismatch, got %v", err)
	}
}

func TestIndexDocument_ResourceNotFound(t *testing.T) {
	s := New()
	ctx := context.Background()

	err := s.IndexDocument(ctx, "srch://nonexistent/doc-1", nil)
	if !errors.Is(err, ErrResourceNotFound) {
		t.Errorf("expected ErrResourceNotFound, got %v", err)
	}
}

func TestIndexDocument_KeyRequired(t *testing.T) {
	s := New()
	search := newMockSearch("articles", articleSpec)
	_ = s.RegisterSearch("srch://articles", search)

	ctx := context.Background()

	err := s.IndexDocument(ctx, "srch://articles", nil)
	if !errors.Is(err, ErrKeyRequired) {
		t.Errorf("expected ErrKeyRequired, got %v", err)
	}
}

func TestDeleteDocument(t *testing.T) {
	s := New()
	search := newMockSearch("articles", articleSpec)
	_ = s.RegisterSearch("srch://articles", search)

	ctx := context.Background()
	search.data["doc-1"] = &grub.AtomicDocument{ID: "doc-1"}

	err := s.DeleteDocument(ctx, "srch://articles/doc-1")
	if err != nil {
		t.Fatalf("DeleteDocument failed: %v", err)
	}

	if _, ok := search.data["doc-1"]; ok {
		t.Error("expected data to be deleted")
	}
}

func TestDeleteDocument_NotFound(t *testing.T) {
	s := New()
	search := newMockSearch("articles", articleSpec)
	_ = s.RegisterSearch("srch://articles", search)

	ctx := context.Background()

	err := s.DeleteDocument(ctx, "srch://articles/doc-1")
	if !errors.Is(err, grub.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteDocument_VariantMismatch(t *testing.T) {
	s := New()
	db := newMockDatabase("users", userSpec)
	_ = s.RegisterDatabase("db://users", db)

	ctx := context.Background()

	err := s.DeleteDocument(ctx, "db://users/123")
	if !errors.Is(err, ErrVariantMismatch) {
		t.Errorf("expected ErrVariantMismatch, got %v", err)
	}
}

func TestDeleteDocument_ResourceNotFound(t *testing.T) {
	s := New()
	ctx := context.Background()

	err := s.DeleteDocument(ctx, "srch://nonexistent/doc-1")
	if !errors.Is(err, ErrResourceNotFound) {
		t.Errorf("expected ErrResourceNotFound, got %v", err)
	}
}

func TestDeleteDocument_KeyRequired(t *testing.T) {
	s := New()
	search := newMockSearch("articles", articleSpec)
	_ = s.RegisterSearch("srch://articles", search)

	ctx := context.Background()

	err := s.DeleteDocument(ctx, "srch://articles")
	if !errors.Is(err, ErrKeyRequired) {
		t.Errorf("expected ErrKeyRequired, got %v", err)
	}
}

func TestDocumentExists(t *testing.T) {
	s := New()
	search := newMockSearch("articles", articleSpec)
	_ = s.RegisterSearch("srch://articles", search)

	ctx := context.Background()
	search.data["doc-1"] = &grub.AtomicDocument{ID: "doc-1"}

	exists, err := s.DocumentExists(ctx, "srch://articles/doc-1")
	if err != nil {
		t.Fatalf("DocumentExists failed: %v", err)
	}
	if !exists {
		t.Error("expected exists=true")
	}

	exists, err = s.DocumentExists(ctx, "srch://articles/doc-2")
	if err != nil {
		t.Fatalf("DocumentExists failed: %v", err)
	}
	if exists {
		t.Error("expected exists=false")
	}
}

func TestDocumentExists_VariantMismatch(t *testing.T) {
	s := New()
	db := newMockDatabase("users", userSpec)
	_ = s.RegisterDatabase("db://users", db)

	ctx := context.Background()

	_, err := s.DocumentExists(ctx, "db://users/123")
	if !errors.Is(err, ErrVariantMismatch) {
		t.Errorf("expected ErrVariantMismatch, got %v", err)
	}
}

func TestDocumentExists_ResourceNotFound(t *testing.T) {
	s := New()
	ctx := context.Background()

	_, err := s.DocumentExists(ctx, "srch://nonexistent/doc-1")
	if !errors.Is(err, ErrResourceNotFound) {
		t.Errorf("expected ErrResourceNotFound, got %v", err)
	}
}

func TestDocumentExists_KeyRequired(t *testing.T) {
	s := New()
	search := newMockSearch("articles", articleSpec)
	_ = s.RegisterSearch("srch://articles", search)

	ctx := context.Background()

	_, err := s.DocumentExists(ctx, "srch://articles")
	if !errors.Is(err, ErrKeyRequired) {
		t.Errorf("expected ErrKeyRequired, got %v", err)
	}
}

func TestSearchDocuments(t *testing.T) {
	s := New()
	search := newMockSearch("articles", articleSpec)
	_ = s.RegisterSearch("srch://articles", search)

	ctx := context.Background()
	search.data["doc-1"] = &grub.AtomicDocument{ID: "doc-1", Content: &atom.Atom{Strings: map[string]string{"Title": "one"}}}
	search.data["doc-2"] = &grub.AtomicDocument{ID: "doc-2", Content: &atom.Atom{Strings: map[string]string{"Title": "two"}}}

	results, err := s.SearchDocuments(ctx, "srch://articles", lucene.NewSearch())
	if err != nil {
		t.Fatalf("SearchDocuments failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestSearchDocuments_VariantMismatch(t *testing.T) {
	s := New()
	db := newMockDatabase("users", userSpec)
	_ = s.RegisterDatabase("db://users", db)

	ctx := context.Background()

	_, err := s.SearchDocuments(ctx, "db://users", lucene.NewSearch())
	if !errors.Is(err, ErrVariantMismatch) {
		t.Errorf("expected ErrVariantMismatch, got %v", err)
	}
}

func TestSearchDocuments_KeyNotExpected(t *testing.T) {
	s := New()
	search := newMockSearch("articles", articleSpec)
	_ = s.RegisterSearch("srch://articles", search)

	ctx := context.Background()

	_, err := s.SearchDocuments(ctx, "srch://articles/doc-1", lucene.NewSearch())
	if !errors.Is(err, ErrKeyNotExpected) {
		t.Errorf("expected ErrKeyNotExpected, got %v", err)
	}
}

func TestSearchDocuments_ResourceNotFound(t *testing.T) {
	s := New()
	ctx := context.Background()

	_, err := s.SearchDocuments(ctx, "srch://nonexistent", lucene.NewSearch())
	if !errors.Is(err, ErrResourceNotFound) {
		t.Errorf("expected ErrResourceNotFound, got %v", err)
	}
}

func TestSearches_InSources(t *testing.T) {
	s := New()
	_ = s.RegisterDatabase("db://users", newMockDatabase("users", userSpec))
	_ = s.RegisterSearch("srch://articles", newMockSearch("articles", articleSpec))

	sources := s.Sources()
	if len(sources) != 2 {
		t.Errorf("expected 2 sources, got %d", len(sources))
	}
}

func TestGetViaSearch(t *testing.T) {
	s := New()
	search := newMockSearch("articles", articleSpec)
	_ = s.RegisterSearch("srch://articles", search)

	ctx := context.Background()
	testContent := &atom.Atom{Strings: map[string]string{"Title": "test"}}
	search.data["doc-1"] = &grub.AtomicDocument{ID: "doc-1", Content: testContent}

	got, err := s.Get(ctx, "srch://articles/doc-1")
	if err != nil {
		t.Fatalf("Get via search failed: %v", err)
	}
	if got.Strings["Title"] != "test" {
		t.Errorf("expected Title=test, got %v", got.Strings["Title"])
	}
}

func TestSetViaSearch_VariantMismatch(t *testing.T) {
	s := New()
	search := newMockSearch("articles", articleSpec)
	_ = s.RegisterSearch("srch://articles", search)

	ctx := context.Background()

	err := s.Set(ctx, "srch://articles/doc-1", &atom.Atom{})
	if !errors.Is(err, ErrVariantMismatch) {
		t.Errorf("expected ErrVariantMismatch, got %v", err)
	}
}

func TestDeleteViaSearch(t *testing.T) {
	s := New()
	search := newMockSearch("articles", articleSpec)
	_ = s.RegisterSearch("srch://articles", search)

	ctx := context.Background()
	search.data["doc-1"] = &grub.AtomicDocument{ID: "doc-1"}

	err := s.Delete(ctx, "srch://articles/doc-1")
	if err != nil {
		t.Fatalf("Delete via search failed: %v", err)
	}
}

func TestExistsViaSearch(t *testing.T) {
	s := New()
	search := newMockSearch("articles", articleSpec)
	_ = s.RegisterSearch("srch://articles", search)

	ctx := context.Background()
	search.data["doc-1"] = &grub.AtomicDocument{ID: "doc-1"}

	exists, err := s.Exists(ctx, "srch://articles/doc-1")
	if err != nil {
		t.Fatalf("Exists via search failed: %v", err)
	}
	if !exists {
		t.Error("expected exists=true")
	}
}
