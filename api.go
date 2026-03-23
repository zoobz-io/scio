// Package scio provides a URI-based data catalog with atomic operations.
// It serves as the authoritative map of data sources, providing topology
// intelligence and type-agnostic access via atoms.
package scio

import (
	"context"
	"time"

	"github.com/zoobz-io/atom"
	"github.com/zoobz-io/edamame"
	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/lucene"
	"github.com/zoobz-io/vecna"
)

// Variant represents a storage type.
type Variant string

const (
	// VariantDatabase represents SQL database storage (db://).
	VariantDatabase Variant = "db"

	// VariantStore represents key-value storage (kv://).
	VariantStore Variant = "kv"

	// VariantBucket represents blob/object storage (bcs://).
	VariantBucket Variant = "bcs"

	// VariantIndex represents vector index storage (idx://).
	VariantIndex Variant = "idx"

	// VariantSearch represents full-text search storage (srch://).
	VariantSearch Variant = "srch"
)

// URI represents a parsed data address.
type URI struct {
	Variant  Variant
	Resource string // table, store, or bucket name
	Key      string // record key or blob path (may be empty for queries)
}

// Resource represents a registered data source.
type Resource struct {
	URI      string
	Variant  Variant
	Name     string
	Spec     atom.Spec
	Metadata Metadata
}

// Metadata holds resource annotations for system use.
type Metadata struct {
	Description string
	Version     string
	Tags        map[string]string
}

// Catalog defines topology introspection operations.
type Catalog interface {
	// Sources returns all registered resources.
	Sources() []Resource

	// Databases returns all db:// resources.
	Databases() []Resource

	// Stores returns all kv:// resources.
	Stores() []Resource

	// Buckets returns all bcs:// resources.
	Buckets() []Resource

	// Indexes returns all idx:// resources.
	Indexes() []Resource

	// Searches returns all srch:// resources.
	Searches() []Resource

	// Spec returns the atom spec for a specific resource.
	Spec(uri string) (atom.Spec, error)

	// FindBySpec returns all resources sharing the given spec (by FQDN).
	FindBySpec(spec atom.Spec) []Resource

	// FindByField returns all resources containing the given field.
	FindByField(field string) []Resource

	// Related returns other resources with the same spec as the given URI.
	Related(uri string) []Resource
}

// Operations defines data access operations.
type Operations interface {
	// Get retrieves an atom at the given URI.
	// Returns ErrNotFound if the key does not exist.
	Get(ctx context.Context, uri string) (*atom.Atom, error)

	// Set stores an atom at the given URI.
	Set(ctx context.Context, uri string, data *atom.Atom) error

	// Delete removes the record at the given URI.
	// Returns ErrNotFound if the key does not exist.
	Delete(ctx context.Context, uri string) error

	// Exists checks whether a record exists at the given URI.
	Exists(ctx context.Context, uri string) (bool, error)

	// Query executes a query statement against a database resource.
	// The URI should reference a db:// resource without a key.
	Query(ctx context.Context, uri string, stmt edamame.QueryStatement, params map[string]any) ([]*atom.Atom, error)

	// Select executes a select statement against a database resource.
	// The URI should reference a db:// resource without a key.
	Select(ctx context.Context, uri string, stmt edamame.SelectStatement, params map[string]any) (*atom.Atom, error)

	// SetWithTTL stores an atom at the given kv:// URI with a TTL.
	// TTL of 0 means no expiration.
	SetWithTTL(ctx context.Context, uri string, data *atom.Atom, ttl time.Duration) error

	// Put stores a blob object at the given bcs:// URI.
	Put(ctx context.Context, uri string, obj *grub.AtomicObject) error

	// NOTE: List operation for bcs:// is deferred pending grub.AtomicBucket interface extension.
}

// Registry defines resource registration operations.
type Registry interface {
	// RegisterDatabase registers an atomic database at the given URI.
	// The URI should be in the form db://table.
	RegisterDatabase(uri string, db grub.AtomicDatabase, opts ...RegistrationOption) error

	// RegisterStore registers an atomic store at the given URI.
	// The URI should be in the form kv://store.
	RegisterStore(uri string, store grub.AtomicStore, opts ...RegistrationOption) error

	// RegisterBucket registers an atomic bucket at the given URI.
	// The URI should be in the form bcs://bucket.
	RegisterBucket(uri string, bucket grub.AtomicBucket, opts ...RegistrationOption) error

	// RegisterIndex registers an atomic index at the given URI.
	// The URI should be in the form idx://index.
	RegisterIndex(uri string, index grub.AtomicIndex, opts ...RegistrationOption) error

	// RegisterSearch registers an atomic search at the given URI.
	// The URI should be in the form srch://index.
	RegisterSearch(uri string, search grub.AtomicSearch, opts ...RegistrationOption) error
}

// IndexOperations defines vector index access operations.
type IndexOperations interface {
	// GetVector retrieves a vector at the given idx:// URI.
	// The URI should include a UUID key component.
	// Returns ErrNotFound if the ID does not exist.
	GetVector(ctx context.Context, uri string) (*grub.AtomicVector, error)

	// UpsertVector stores a vector with metadata at the given idx:// URI.
	// The URI should include a UUID key component.
	UpsertVector(ctx context.Context, uri string, vector []float32, metadata *atom.Atom) error

	// DeleteVector removes the vector at the given idx:// URI.
	// The URI should include a UUID key component.
	// Returns ErrNotFound if the ID does not exist.
	DeleteVector(ctx context.Context, uri string) error

	// VectorExists checks whether a vector exists at the given idx:// URI.
	// The URI should include a UUID key component.
	VectorExists(ctx context.Context, uri string) (bool, error)

	// SearchVectors performs similarity search against an idx:// resource.
	// The URI should reference the index without a key component.
	// Returns the k nearest neighbors, optionally filtered by metadata.
	SearchVectors(ctx context.Context, uri string, vector []float32, k int, filter *atom.Atom) ([]grub.AtomicVector, error)

	// QueryVectors performs similarity search with vecna filter support.
	// The URI should reference the index without a key component.
	QueryVectors(ctx context.Context, uri string, vector []float32, k int, filter *vecna.Filter) ([]grub.AtomicVector, error)

	// FilterVectors returns vectors matching the metadata filter without similarity search.
	// The URI should reference the index without a key component.
	FilterVectors(ctx context.Context, uri string, filter *vecna.Filter, limit int) ([]grub.AtomicVector, error)
}

// SearchOperations defines full-text search access operations.
type SearchOperations interface {
	// GetDocument retrieves a document at the given srch:// URI.
	// The URI should include a document ID key component.
	// Returns ErrNotFound if the ID does not exist.
	GetDocument(ctx context.Context, uri string) (*grub.AtomicDocument, error)

	// IndexDocument stores a document with atomized content at the given srch:// URI.
	// The URI should include a document ID key component.
	IndexDocument(ctx context.Context, uri string, data *atom.Atom) error

	// DeleteDocument removes the document at the given srch:// URI.
	// The URI should include a document ID key component.
	// Returns ErrNotFound if the ID does not exist.
	DeleteDocument(ctx context.Context, uri string) error

	// DocumentExists checks whether a document exists at the given srch:// URI.
	// The URI should include a document ID key component.
	DocumentExists(ctx context.Context, uri string) (bool, error)

	// SearchDocuments performs a full-text search against a srch:// resource.
	// The URI should reference the search index without a key component.
	SearchDocuments(ctx context.Context, uri string, search *lucene.Search) ([]grub.AtomicDocument, error)
}
