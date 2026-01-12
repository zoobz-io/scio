package scio

import (
	"context"
	"time"

	"github.com/zoobzio/atom"
	"github.com/zoobzio/edamame"
	"github.com/zoobzio/grub"
)

// Get retrieves an atom at the given URI.
// Returns ErrNotFound if the key does not exist.
// Returns ErrResourceNotFound if the resource is not registered.
// Returns ErrKeyRequired if no key is provided.
func (s *Scio) Get(ctx context.Context, uri string) (*atom.Atom, error) {
	parsed, err := ParseURI(uri)
	if err != nil {
		return nil, err
	}

	if parsed.Key == "" {
		return nil, ErrKeyRequired
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()

	switch parsed.Variant {
	case VariantDatabase:
		db, ok := s.getDatabase(resourceURI)
		if !ok {
			return nil, ErrResourceNotFound
		}
		return db.Get(ctx, parsed.Key)

	case VariantStore:
		store, ok := s.getStore(resourceURI)
		if !ok {
			return nil, ErrResourceNotFound
		}
		return store.Get(ctx, parsed.Key)

	case VariantBucket:
		bucket, ok := s.getBucket(resourceURI)
		if !ok {
			return nil, ErrResourceNotFound
		}
		obj, err := bucket.Get(ctx, parsed.Key)
		if err != nil {
			return nil, err
		}
		return obj.Data, nil

	default:
		return nil, ErrUnknownVariant
	}
}

// Set stores an atom at the given URI.
// Returns ErrResourceNotFound if the resource is not registered.
// Returns ErrKeyRequired if no key is provided.
// For kv:// resources, use SetWithTTL to specify expiration.
func (s *Scio) Set(ctx context.Context, uri string, data *atom.Atom) error {
	parsed, err := ParseURI(uri)
	if err != nil {
		return err
	}

	if parsed.Key == "" {
		return ErrKeyRequired
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()

	switch parsed.Variant {
	case VariantDatabase:
		db, ok := s.getDatabase(resourceURI)
		if !ok {
			return ErrResourceNotFound
		}
		return db.Set(ctx, parsed.Key, data)

	case VariantStore:
		store, ok := s.getStore(resourceURI)
		if !ok {
			return ErrResourceNotFound
		}
		return store.Set(ctx, parsed.Key, data, 0)

	case VariantBucket:
		return ErrVariantMismatch // Use Put for buckets

	default:
		return ErrUnknownVariant
	}
}

// SetWithTTL stores an atom at the given kv:// URI with a TTL.
// TTL of 0 means no expiration.
// Returns ErrVariantMismatch if the URI is not a kv:// resource.
func (s *Scio) SetWithTTL(ctx context.Context, uri string, data *atom.Atom, ttl time.Duration) error {
	parsed, err := ParseURI(uri)
	if err != nil {
		return err
	}

	if parsed.Variant != VariantStore {
		return ErrVariantMismatch
	}

	if parsed.Key == "" {
		return ErrKeyRequired
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()
	store, ok := s.getStore(resourceURI)
	if !ok {
		return ErrResourceNotFound
	}

	return store.Set(ctx, parsed.Key, data, ttl)
}

// Delete removes the record at the given URI.
// Returns ErrNotFound if the key does not exist.
// Returns ErrResourceNotFound if the resource is not registered.
func (s *Scio) Delete(ctx context.Context, uri string) error {
	parsed, err := ParseURI(uri)
	if err != nil {
		return err
	}

	if parsed.Key == "" {
		return ErrKeyRequired
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()

	switch parsed.Variant {
	case VariantDatabase:
		db, ok := s.getDatabase(resourceURI)
		if !ok {
			return ErrResourceNotFound
		}
		return db.Delete(ctx, parsed.Key)

	case VariantStore:
		store, ok := s.getStore(resourceURI)
		if !ok {
			return ErrResourceNotFound
		}
		return store.Delete(ctx, parsed.Key)

	case VariantBucket:
		bucket, ok := s.getBucket(resourceURI)
		if !ok {
			return ErrResourceNotFound
		}
		return bucket.Delete(ctx, parsed.Key)

	default:
		return ErrUnknownVariant
	}
}

// Exists checks whether a record exists at the given URI.
// Returns ErrResourceNotFound if the resource is not registered.
func (s *Scio) Exists(ctx context.Context, uri string) (bool, error) {
	parsed, err := ParseURI(uri)
	if err != nil {
		return false, err
	}

	if parsed.Key == "" {
		return false, ErrKeyRequired
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()

	switch parsed.Variant {
	case VariantDatabase:
		db, ok := s.getDatabase(resourceURI)
		if !ok {
			return false, ErrResourceNotFound
		}
		return db.Exists(ctx, parsed.Key)

	case VariantStore:
		store, ok := s.getStore(resourceURI)
		if !ok {
			return false, ErrResourceNotFound
		}
		return store.Exists(ctx, parsed.Key)

	case VariantBucket:
		bucket, ok := s.getBucket(resourceURI)
		if !ok {
			return false, ErrResourceNotFound
		}
		return bucket.Exists(ctx, parsed.Key)

	default:
		return false, ErrUnknownVariant
	}
}

// Query executes a query statement against a database resource.
// The URI should reference a db:// resource without a key component.
// Returns ErrVariantMismatch if the URI is not a db:// resource.
// Returns ErrKeyNotExpected if a key component is provided.
func (s *Scio) Query(ctx context.Context, uri string, stmt edamame.QueryStatement, params map[string]any) ([]*atom.Atom, error) {
	parsed, err := ParseURI(uri)
	if err != nil {
		return nil, err
	}

	if parsed.Variant != VariantDatabase {
		return nil, ErrVariantMismatch
	}

	if parsed.Key != "" {
		return nil, ErrKeyNotExpected
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()
	db, ok := s.getDatabase(resourceURI)
	if !ok {
		return nil, ErrResourceNotFound
	}

	return db.Query(ctx, stmt, params)
}

// Select executes a select statement against a database resource.
// The URI should reference a db:// resource without a key component.
// Returns ErrVariantMismatch if the URI is not a db:// resource.
// Returns ErrKeyNotExpected if a key component is provided.
func (s *Scio) Select(ctx context.Context, uri string, stmt edamame.SelectStatement, params map[string]any) (*atom.Atom, error) {
	parsed, err := ParseURI(uri)
	if err != nil {
		return nil, err
	}

	if parsed.Variant != VariantDatabase {
		return nil, ErrVariantMismatch
	}

	if parsed.Key != "" {
		return nil, ErrKeyNotExpected
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()
	db, ok := s.getDatabase(resourceURI)
	if !ok {
		return nil, ErrResourceNotFound
	}

	return db.Select(ctx, stmt, params)
}

// Put stores a blob object at the given bcs:// URI.
// Returns ErrVariantMismatch if the URI is not a bcs:// resource.
func (s *Scio) Put(ctx context.Context, uri string, obj *grub.AtomicObject) error {
	parsed, err := ParseURI(uri)
	if err != nil {
		return err
	}

	if parsed.Variant != VariantBucket {
		return ErrVariantMismatch
	}

	if parsed.Key == "" {
		return ErrKeyRequired
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()
	bucket, ok := s.getBucket(resourceURI)
	if !ok {
		return ErrResourceNotFound
	}

	return bucket.Put(ctx, parsed.Key, obj)
}

// NOTE: List operation for bcs:// is deferred pending grub.AtomicBucket interface extension.
