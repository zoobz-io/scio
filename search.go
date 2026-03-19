package scio

import (
	"context"

	"github.com/zoobzio/atom"
	"github.com/zoobzio/grub"
	"github.com/zoobzio/lucene"
)

// GetDocument retrieves a document at the given srch:// URI.
// The URI should include a document ID key component.
// Returns ErrNotFound if the ID does not exist.
// Returns ErrResourceNotFound if the resource is not registered.
// Returns ErrKeyRequired if no key is provided.
func (s *Scio) GetDocument(ctx context.Context, uri string) (*grub.AtomicDocument, error) {
	parsed, err := ParseURI(uri)
	if err != nil {
		return nil, err
	}

	if parsed.Variant != VariantSearch {
		return nil, ErrVariantMismatch
	}

	if parsed.Key == "" {
		return nil, ErrKeyRequired
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()
	search, ok := s.getSearch(resourceURI)
	if !ok {
		return nil, ErrResourceNotFound
	}

	return search.Get(ctx, parsed.Key)
}

// IndexDocument stores a document with atomized content at the given srch:// URI.
// The URI should include a document ID key component.
// Returns ErrResourceNotFound if the resource is not registered.
// Returns ErrKeyRequired if no key is provided.
func (s *Scio) IndexDocument(ctx context.Context, uri string, data *atom.Atom) error {
	parsed, err := ParseURI(uri)
	if err != nil {
		return err
	}

	if parsed.Variant != VariantSearch {
		return ErrVariantMismatch
	}

	if parsed.Key == "" {
		return ErrKeyRequired
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()
	search, ok := s.getSearch(resourceURI)
	if !ok {
		return ErrResourceNotFound
	}

	return search.IndexDoc(ctx, parsed.Key, data)
}

// DeleteDocument removes the document at the given srch:// URI.
// The URI should include a document ID key component.
// Returns ErrNotFound if the ID does not exist.
// Returns ErrResourceNotFound if the resource is not registered.
func (s *Scio) DeleteDocument(ctx context.Context, uri string) error {
	parsed, err := ParseURI(uri)
	if err != nil {
		return err
	}

	if parsed.Variant != VariantSearch {
		return ErrVariantMismatch
	}

	if parsed.Key == "" {
		return ErrKeyRequired
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()
	search, ok := s.getSearch(resourceURI)
	if !ok {
		return ErrResourceNotFound
	}

	return search.Delete(ctx, parsed.Key)
}

// DocumentExists checks whether a document exists at the given srch:// URI.
// The URI should include a document ID key component.
// Returns ErrResourceNotFound if the resource is not registered.
func (s *Scio) DocumentExists(ctx context.Context, uri string) (bool, error) {
	parsed, err := ParseURI(uri)
	if err != nil {
		return false, err
	}

	if parsed.Variant != VariantSearch {
		return false, ErrVariantMismatch
	}

	if parsed.Key == "" {
		return false, ErrKeyRequired
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()
	search, ok := s.getSearch(resourceURI)
	if !ok {
		return false, ErrResourceNotFound
	}

	return search.Exists(ctx, parsed.Key)
}

// SearchDocuments performs a full-text search against a srch:// resource.
// The URI should reference the search index without a key component.
// Returns ErrResourceNotFound if the resource is not registered.
// Returns ErrKeyNotExpected if a key component is provided.
func (s *Scio) SearchDocuments(ctx context.Context, uri string, search *lucene.Search) ([]grub.AtomicDocument, error) {
	parsed, err := ParseURI(uri)
	if err != nil {
		return nil, err
	}

	if parsed.Variant != VariantSearch {
		return nil, ErrVariantMismatch
	}

	if parsed.Key != "" {
		return nil, ErrKeyNotExpected
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()
	srch, ok := s.getSearch(resourceURI)
	if !ok {
		return nil, ErrResourceNotFound
	}

	return srch.Search(ctx, search)
}
