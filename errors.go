package scio

import (
	"errors"

	"github.com/zoobzio/grub"
)

// Semantic errors for scio operations.
var (
	// ErrInvalidURI indicates the URI could not be parsed.
	ErrInvalidURI = errors.New("invalid URI")

	// ErrUnknownVariant indicates an unrecognized URI scheme.
	ErrUnknownVariant = errors.New("unknown variant")

	// ErrResourceNotFound indicates the requested resource is not registered.
	ErrResourceNotFound = errors.New("resource not found")

	// ErrResourceExists indicates a resource is already registered at the URI.
	ErrResourceExists = errors.New("resource already exists")

	// ErrVariantMismatch indicates an operation was attempted on the wrong variant.
	ErrVariantMismatch = errors.New("variant mismatch")

	// ErrKeyRequired indicates a key is required but was not provided.
	ErrKeyRequired = errors.New("key required")

	// ErrKeyNotExpected indicates a key was provided but is not used by the operation.
	ErrKeyNotExpected = errors.New("key not expected")
)

// Re-export grub errors for convenience.
var (
	ErrNotFound        = grub.ErrNotFound
	ErrDuplicate       = grub.ErrDuplicate
	ErrConflict        = grub.ErrConflict
	ErrConstraint      = grub.ErrConstraint
	ErrInvalidKey      = grub.ErrInvalidKey
	ErrReadOnly        = grub.ErrReadOnly
	ErrTTLNotSupported = grub.ErrTTLNotSupported
)
