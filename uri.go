package scio

import (
	"strings"
)

// ParseURI parses a raw URI string into its components.
// Supported formats:
//   - db://table/key     -> Variant: db, Resource: table, Key: key
//   - kv://store/key     -> Variant: kv, Resource: store, Key: key
//   - bcs://bucket/path  -> Variant: bcs, Resource: bucket, Key: path (full remainder)
func ParseURI(raw string) (*URI, error) {
	// Find scheme separator
	schemeEnd := strings.Index(raw, "://")
	if schemeEnd == -1 {
		return nil, ErrInvalidURI
	}

	scheme := raw[:schemeEnd]
	remainder := raw[schemeEnd+3:]

	if remainder == "" {
		return nil, ErrInvalidURI
	}

	variant, err := parseVariant(scheme)
	if err != nil {
		return nil, err
	}

	resource, key := parsePath(variant, remainder)
	if resource == "" {
		return nil, ErrInvalidURI
	}

	return &URI{
		Variant:  variant,
		Resource: resource,
		Key:      key,
	}, nil
}

// parseVariant converts a scheme string to a Variant.
func parseVariant(scheme string) (Variant, error) {
	switch scheme {
	case "db":
		return VariantDatabase, nil
	case "kv":
		return VariantStore, nil
	case "bcs":
		return VariantBucket, nil
	case "idx":
		return VariantIndex, nil
	default:
		return "", ErrUnknownVariant
	}
}

// parsePath extracts resource and key from the path based on variant.
func parsePath(variant Variant, path string) (resource, key string) {
	switch variant {
	case VariantDatabase, VariantStore, VariantIndex:
		// db://table/key or kv://store/key or idx://index/uuid
		// First segment is resource, second is key
		parts := strings.SplitN(path, "/", 2)
		resource = parts[0]
		if len(parts) > 1 {
			key = parts[1]
		}

	case VariantBucket:
		// bcs://bucket/path/to/file
		// First segment is bucket, remainder is full path
		parts := strings.SplitN(path, "/", 2)
		resource = parts[0]
		if len(parts) > 1 {
			key = parts[1]
		}
	}

	return resource, key
}

// String returns the URI as a string.
func (u *URI) String() string {
	if u.Key == "" {
		return string(u.Variant) + "://" + u.Resource
	}
	return string(u.Variant) + "://" + u.Resource + "/" + u.Key
}

// ResourceURI returns the URI without the key component.
func (u *URI) ResourceURI() string {
	return string(u.Variant) + "://" + u.Resource
}
