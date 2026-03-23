package scio

import (
	"sync"

	"github.com/zoobz-io/atom"
	"github.com/zoobz-io/grub"
)

// Scio is the central data catalog and access point.
type Scio struct {
	databases map[string]grub.AtomicDatabase
	stores    map[string]grub.AtomicStore
	buckets   map[string]grub.AtomicBucket
	indexes   map[string]grub.AtomicIndex
	searches  map[string]grub.AtomicSearch
	resources map[string]*Resource
	specs     map[string][]*Resource // FQDN -> resources with that spec
	mu        sync.RWMutex
}

// New creates a new Scio instance.
func New() *Scio {
	return &Scio{
		databases: make(map[string]grub.AtomicDatabase),
		stores:    make(map[string]grub.AtomicStore),
		buckets:   make(map[string]grub.AtomicBucket),
		indexes:   make(map[string]grub.AtomicIndex),
		searches:  make(map[string]grub.AtomicSearch),
		resources: make(map[string]*Resource),
		specs:     make(map[string][]*Resource),
	}
}

// RegisterDatabase registers an atomic database at the given URI.
// The URI should be in the form db://table.
func (s *Scio) RegisterDatabase(uri string, db grub.AtomicDatabase, opts ...RegistrationOption) error {
	parsed, err := ParseURI(uri)
	if err != nil {
		return err
	}

	if parsed.Variant != VariantDatabase {
		return ErrVariantMismatch
	}

	resourceURI := parsed.ResourceURI()
	cfg := applyOptions(opts)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.resources[resourceURI]; exists {
		return ErrResourceExists
	}

	spec := db.Spec()
	resource := &Resource{
		URI:      resourceURI,
		Variant:  VariantDatabase,
		Name:     parsed.Resource,
		Spec:     spec,
		Metadata: cfg.metadata,
	}

	s.databases[resourceURI] = db
	s.resources[resourceURI] = resource
	s.trackSpec(spec, resource)

	return nil
}

// RegisterStore registers an atomic store at the given URI.
// The URI should be in the form kv://store.
func (s *Scio) RegisterStore(uri string, store grub.AtomicStore, opts ...RegistrationOption) error {
	parsed, err := ParseURI(uri)
	if err != nil {
		return err
	}

	if parsed.Variant != VariantStore {
		return ErrVariantMismatch
	}

	resourceURI := parsed.ResourceURI()
	cfg := applyOptions(opts)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.resources[resourceURI]; exists {
		return ErrResourceExists
	}

	spec := store.Spec()
	resource := &Resource{
		URI:      resourceURI,
		Variant:  VariantStore,
		Name:     parsed.Resource,
		Spec:     spec,
		Metadata: cfg.metadata,
	}

	s.stores[resourceURI] = store
	s.resources[resourceURI] = resource
	s.trackSpec(spec, resource)

	return nil
}

// RegisterBucket registers an atomic bucket at the given URI.
// The URI should be in the form bcs://bucket.
func (s *Scio) RegisterBucket(uri string, bucket grub.AtomicBucket, opts ...RegistrationOption) error {
	parsed, err := ParseURI(uri)
	if err != nil {
		return err
	}

	if parsed.Variant != VariantBucket {
		return ErrVariantMismatch
	}

	resourceURI := parsed.ResourceURI()
	cfg := applyOptions(opts)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.resources[resourceURI]; exists {
		return ErrResourceExists
	}

	spec := bucket.Spec()
	resource := &Resource{
		URI:      resourceURI,
		Variant:  VariantBucket,
		Name:     parsed.Resource,
		Spec:     spec,
		Metadata: cfg.metadata,
	}

	s.buckets[resourceURI] = bucket
	s.resources[resourceURI] = resource
	s.trackSpec(spec, resource)

	return nil
}

// trackSpec adds a resource to the spec tracking map.
// Caller must hold the write lock.
func (s *Scio) trackSpec(spec atom.Spec, resource *Resource) {
	fqdn := spec.FQDN
	s.specs[fqdn] = append(s.specs[fqdn], resource)
}

// getDatabase retrieves a database by resource URI.
// Caller must hold at least a read lock.
func (s *Scio) getDatabase(resourceURI string) (grub.AtomicDatabase, bool) {
	db, ok := s.databases[resourceURI]
	return db, ok
}

// getStore retrieves a store by resource URI.
// Caller must hold at least a read lock.
func (s *Scio) getStore(resourceURI string) (grub.AtomicStore, bool) {
	store, ok := s.stores[resourceURI]
	return store, ok
}

// getBucket retrieves a bucket by resource URI.
// Caller must hold at least a read lock.
func (s *Scio) getBucket(resourceURI string) (grub.AtomicBucket, bool) {
	bucket, ok := s.buckets[resourceURI]
	return bucket, ok
}

// RegisterIndex registers an atomic index at the given URI.
// The URI should be in the form idx://index.
func (s *Scio) RegisterIndex(uri string, index grub.AtomicIndex, opts ...RegistrationOption) error {
	parsed, err := ParseURI(uri)
	if err != nil {
		return err
	}

	if parsed.Variant != VariantIndex {
		return ErrVariantMismatch
	}

	resourceURI := parsed.ResourceURI()
	cfg := applyOptions(opts)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.resources[resourceURI]; exists {
		return ErrResourceExists
	}

	spec := index.Spec()
	resource := &Resource{
		URI:      resourceURI,
		Variant:  VariantIndex,
		Name:     parsed.Resource,
		Spec:     spec,
		Metadata: cfg.metadata,
	}

	s.indexes[resourceURI] = index
	s.resources[resourceURI] = resource
	s.trackSpec(spec, resource)

	return nil
}

// getIndex retrieves an index by resource URI.
// Caller must hold at least a read lock.
func (s *Scio) getIndex(resourceURI string) (grub.AtomicIndex, bool) {
	index, ok := s.indexes[resourceURI]
	return index, ok
}

// RegisterSearch registers an atomic search at the given URI.
// The URI should be in the form srch://index.
func (s *Scio) RegisterSearch(uri string, search grub.AtomicSearch, opts ...RegistrationOption) error {
	parsed, err := ParseURI(uri)
	if err != nil {
		return err
	}

	if parsed.Variant != VariantSearch {
		return ErrVariantMismatch
	}

	resourceURI := parsed.ResourceURI()
	cfg := applyOptions(opts)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.resources[resourceURI]; exists {
		return ErrResourceExists
	}

	spec := search.Spec()
	resource := &Resource{
		URI:      resourceURI,
		Variant:  VariantSearch,
		Name:     parsed.Resource,
		Spec:     spec,
		Metadata: cfg.metadata,
	}

	s.searches[resourceURI] = search
	s.resources[resourceURI] = resource
	s.trackSpec(spec, resource)

	return nil
}

// getSearch retrieves a search by resource URI.
// Caller must hold at least a read lock.
func (s *Scio) getSearch(resourceURI string) (grub.AtomicSearch, bool) {
	search, ok := s.searches[resourceURI]
	return search, ok
}
