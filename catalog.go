package scio

import (
	"github.com/zoobzio/atom"
)

// Sources returns all registered resources.
func (s *Scio) Sources() []Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Resource, 0, len(s.resources))
	for _, r := range s.resources {
		result = append(result, *r)
	}
	return result
}

// Databases returns all db:// resources.
func (s *Scio) Databases() []Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Resource, 0, len(s.databases))
	for uri := range s.databases {
		if r, ok := s.resources[uri]; ok {
			result = append(result, *r)
		}
	}
	return result
}

// Stores returns all kv:// resources.
func (s *Scio) Stores() []Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Resource, 0, len(s.stores))
	for uri := range s.stores {
		if r, ok := s.resources[uri]; ok {
			result = append(result, *r)
		}
	}
	return result
}

// Buckets returns all bcs:// resources.
func (s *Scio) Buckets() []Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Resource, 0, len(s.buckets))
	for uri := range s.buckets {
		if r, ok := s.resources[uri]; ok {
			result = append(result, *r)
		}
	}
	return result
}

// Indexes returns all idx:// resources.
func (s *Scio) Indexes() []Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Resource, 0, len(s.indexes))
	for uri := range s.indexes {
		if r, ok := s.resources[uri]; ok {
			result = append(result, *r)
		}
	}
	return result
}

// Spec returns the atom spec for a specific resource.
// Returns ErrResourceNotFound if the resource is not registered.
func (s *Scio) Spec(uri string) (atom.Spec, error) {
	parsed, err := ParseURI(uri)
	if err != nil {
		return atom.Spec{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()
	r, ok := s.resources[resourceURI]
	if !ok {
		return atom.Spec{}, ErrResourceNotFound
	}

	return r.Spec, nil
}

// FindBySpec returns all resources sharing the given spec (by FQDN).
func (s *Scio) FindBySpec(spec atom.Spec) []Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resources, ok := s.specs[spec.FQDN]
	if !ok {
		return []Resource{}
	}

	result := make([]Resource, len(resources))
	for i, r := range resources {
		result[i] = *r
	}
	return result
}

// FindByField returns all resources containing the given field.
func (s *Scio) FindByField(field string) []Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []Resource
	for _, r := range s.resources {
		if hasField(r.Spec, field) {
			result = append(result, *r)
		}
	}
	return result
}

// Related returns other resources with the same spec as the given URI.
// The resource at the given URI is excluded from the results.
func (s *Scio) Related(uri string) []Resource {
	parsed, err := ParseURI(uri)
	if err != nil {
		return []Resource{}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()
	r, ok := s.resources[resourceURI]
	if !ok {
		return []Resource{}
	}

	resources, ok := s.specs[r.Spec.FQDN]
	if !ok {
		return []Resource{}
	}

	result := make([]Resource, 0, len(resources))
	for _, related := range resources {
		if related.URI != resourceURI {
			result = append(result, *related)
		}
	}
	return result
}

// Resource returns the resource metadata for a specific URI.
// Returns nil if the resource is not registered.
func (s *Scio) Resource(uri string) *Resource {
	parsed, err := ParseURI(uri)
	if err != nil {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resourceURI := parsed.ResourceURI()
	r, ok := s.resources[resourceURI]
	if !ok {
		return nil
	}

	// Return a copy to prevent mutation
	result := *r
	return &result
}

// hasField checks if a spec contains a field with the given name.
func hasField(spec atom.Spec, field string) bool {
	for _, f := range spec.Fields {
		if f.Name == field {
			return true
		}
	}
	return false
}
