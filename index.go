package scio

import (
	"context"

	"github.com/google/uuid"
	"github.com/zoobz-io/atom"
	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/vecna"
)

// GetVector retrieves a vector at the given idx:// URI.
// The URI should include a UUID key component.
// Returns ErrNotFound if the ID does not exist.
// Returns ErrResourceNotFound if the resource is not registered.
// Returns ErrKeyRequired if no key is provided.
func (s *Scio) GetVector(ctx context.Context, uri string) (*grub.AtomicVector, error) {
	parsed, err := ParseURI(uri)
	if err != nil {
		return nil, err
	}

	if parsed.Variant != VariantIndex {
		return nil, ErrVariantMismatch
	}

	if parsed.Key == "" {
		return nil, ErrKeyRequired
	}

	id, err := uuid.Parse(parsed.Key)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()
	index, ok := s.getIndex(resourceURI)
	if !ok {
		return nil, ErrResourceNotFound
	}

	return index.Get(ctx, id)
}

// UpsertVector stores a vector with metadata at the given idx:// URI.
// The URI should include a UUID key component.
// Returns ErrResourceNotFound if the resource is not registered.
// Returns ErrKeyRequired if no key is provided.
func (s *Scio) UpsertVector(ctx context.Context, uri string, vector []float32, metadata *atom.Atom) error {
	parsed, err := ParseURI(uri)
	if err != nil {
		return err
	}

	if parsed.Variant != VariantIndex {
		return ErrVariantMismatch
	}

	if parsed.Key == "" {
		return ErrKeyRequired
	}

	id, err := uuid.Parse(parsed.Key)
	if err != nil {
		return ErrInvalidUUID
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()
	index, ok := s.getIndex(resourceURI)
	if !ok {
		return ErrResourceNotFound
	}

	return index.Upsert(ctx, id, vector, metadata)
}

// DeleteVector removes the vector at the given idx:// URI.
// The URI should include a UUID key component.
// Returns ErrNotFound if the ID does not exist.
// Returns ErrResourceNotFound if the resource is not registered.
func (s *Scio) DeleteVector(ctx context.Context, uri string) error {
	parsed, err := ParseURI(uri)
	if err != nil {
		return err
	}

	if parsed.Variant != VariantIndex {
		return ErrVariantMismatch
	}

	if parsed.Key == "" {
		return ErrKeyRequired
	}

	id, err := uuid.Parse(parsed.Key)
	if err != nil {
		return ErrInvalidUUID
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()
	index, ok := s.getIndex(resourceURI)
	if !ok {
		return ErrResourceNotFound
	}

	return index.Delete(ctx, id)
}

// VectorExists checks whether a vector exists at the given idx:// URI.
// The URI should include a UUID key component.
// Returns ErrResourceNotFound if the resource is not registered.
func (s *Scio) VectorExists(ctx context.Context, uri string) (bool, error) {
	parsed, err := ParseURI(uri)
	if err != nil {
		return false, err
	}

	if parsed.Variant != VariantIndex {
		return false, ErrVariantMismatch
	}

	if parsed.Key == "" {
		return false, ErrKeyRequired
	}

	id, err := uuid.Parse(parsed.Key)
	if err != nil {
		return false, ErrInvalidUUID
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()
	index, ok := s.getIndex(resourceURI)
	if !ok {
		return false, ErrResourceNotFound
	}

	return index.Exists(ctx, id)
}

// SearchVectors performs similarity search against an idx:// resource.
// The URI should reference the index without a key component.
// Returns the k nearest neighbors, optionally filtered by metadata.
// Returns ErrResourceNotFound if the resource is not registered.
// Returns ErrKeyNotExpected if a key component is provided.
func (s *Scio) SearchVectors(ctx context.Context, uri string, vector []float32, k int, filter *atom.Atom) ([]grub.AtomicVector, error) {
	parsed, err := ParseURI(uri)
	if err != nil {
		return nil, err
	}

	if parsed.Variant != VariantIndex {
		return nil, ErrVariantMismatch
	}

	if parsed.Key != "" {
		return nil, ErrKeyNotExpected
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()
	index, ok := s.getIndex(resourceURI)
	if !ok {
		return nil, ErrResourceNotFound
	}

	return index.Search(ctx, vector, k, filter)
}

// QueryVectors performs similarity search with vecna filter support.
// The URI should reference the index without a key component.
// Returns ErrResourceNotFound if the resource is not registered.
// Returns ErrKeyNotExpected if a key component is provided.
func (s *Scio) QueryVectors(ctx context.Context, uri string, vector []float32, k int, filter *vecna.Filter) ([]grub.AtomicVector, error) {
	parsed, err := ParseURI(uri)
	if err != nil {
		return nil, err
	}

	if parsed.Variant != VariantIndex {
		return nil, ErrVariantMismatch
	}

	if parsed.Key != "" {
		return nil, ErrKeyNotExpected
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()
	index, ok := s.getIndex(resourceURI)
	if !ok {
		return nil, ErrResourceNotFound
	}

	return index.Query(ctx, vector, k, filter)
}

// FilterVectors returns vectors matching the metadata filter without similarity search.
// The URI should reference the index without a key component.
// Returns ErrResourceNotFound if the resource is not registered.
// Returns ErrKeyNotExpected if a key component is provided.
func (s *Scio) FilterVectors(ctx context.Context, uri string, filter *vecna.Filter, limit int) ([]grub.AtomicVector, error) {
	parsed, err := ParseURI(uri)
	if err != nil {
		return nil, err
	}

	if parsed.Variant != VariantIndex {
		return nil, ErrVariantMismatch
	}

	if parsed.Key != "" {
		return nil, ErrKeyNotExpected
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()
	index, ok := s.getIndex(resourceURI)
	if !ok {
		return nil, ErrResourceNotFound
	}

	return index.Filter(ctx, filter, limit)
}
