package scio

import (
	"errors"
	"testing"
)

func TestParseURI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     *URI
		wantErr  error
	}{
		{
			name:  "database with key",
			input: "db://users/123",
			want: &URI{
				Variant:  VariantDatabase,
				Resource: "users",
				Key:      "123",
			},
		},
		{
			name:  "database without key",
			input: "db://users",
			want: &URI{
				Variant:  VariantDatabase,
				Resource: "users",
				Key:      "",
			},
		},
		{
			name:  "store with key",
			input: "kv://sessions/abc-123",
			want: &URI{
				Variant:  VariantStore,
				Resource: "sessions",
				Key:      "abc-123",
			},
		},
		{
			name:  "store without key",
			input: "kv://cache",
			want: &URI{
				Variant:  VariantStore,
				Resource: "cache",
				Key:      "",
			},
		},
		{
			name:  "bucket with simple path",
			input: "bcs://documents/report.pdf",
			want: &URI{
				Variant:  VariantBucket,
				Resource: "documents",
				Key:      "report.pdf",
			},
		},
		{
			name:  "bucket with nested path",
			input: "bcs://documents/reports/2024/q4/summary.pdf",
			want: &URI{
				Variant:  VariantBucket,
				Resource: "documents",
				Key:      "reports/2024/q4/summary.pdf",
			},
		},
		{
			name:  "bucket without path",
			input: "bcs://assets",
			want: &URI{
				Variant:  VariantBucket,
				Resource: "assets",
				Key:      "",
			},
		},
		{
			name:    "missing scheme separator",
			input:   "db/users/123",
			wantErr: ErrInvalidURI,
		},
		{
			name:    "empty path",
			input:   "db://",
			wantErr: ErrInvalidURI,
		},
		{
			name:    "unknown variant",
			input:   "sql://users/123",
			wantErr: ErrUnknownVariant,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: ErrInvalidURI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseURI(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Variant != tt.want.Variant {
				t.Errorf("Variant = %v, want %v", got.Variant, tt.want.Variant)
			}
			if got.Resource != tt.want.Resource {
				t.Errorf("Resource = %v, want %v", got.Resource, tt.want.Resource)
			}
			if got.Key != tt.want.Key {
				t.Errorf("Key = %v, want %v", got.Key, tt.want.Key)
			}
		})
	}
}

func TestURI_String(t *testing.T) {
	tests := []struct {
		name string
		uri  URI
		want string
	}{
		{
			name: "database with key",
			uri:  URI{Variant: VariantDatabase, Resource: "users", Key: "123"},
			want: "db://users/123",
		},
		{
			name: "database without key",
			uri:  URI{Variant: VariantDatabase, Resource: "users", Key: ""},
			want: "db://users",
		},
		{
			name: "bucket with nested path",
			uri:  URI{Variant: VariantBucket, Resource: "docs", Key: "a/b/c.pdf"},
			want: "bcs://docs/a/b/c.pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.uri.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURI_ResourceURI(t *testing.T) {
	tests := []struct {
		name string
		uri  URI
		want string
	}{
		{
			name: "strips key from database",
			uri:  URI{Variant: VariantDatabase, Resource: "users", Key: "123"},
			want: "db://users",
		},
		{
			name: "strips path from bucket",
			uri:  URI{Variant: VariantBucket, Resource: "docs", Key: "a/b/c.pdf"},
			want: "bcs://docs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.uri.ResourceURI(); got != tt.want {
				t.Errorf("ResourceURI() = %v, want %v", got, tt.want)
			}
		})
	}
}
