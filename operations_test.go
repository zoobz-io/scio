package scio

import (
	"context"
	"errors"
	"testing"

	"github.com/zoobzio/atom"
	"github.com/zoobzio/grub"
)

func TestGet_Database(t *testing.T) {
	s := New()
	db := newMockDatabase("users", userSpec)
	_ = s.RegisterDatabase("db://users", db)

	ctx := context.Background()

	// Set data directly in mock
	testAtom := &atom.Atom{Strings: map[string]string{"Name": "Alice"}}
	db.data["123"] = testAtom

	got, err := s.Get(ctx, "db://users/123")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Strings["Name"] != "Alice" {
		t.Errorf("expected Name=Alice, got %v", got.Strings["Name"])
	}
}

func TestGet_Store(t *testing.T) {
	s := New()
	store := newMockStore(sessionSpec)
	_ = s.RegisterStore("kv://sessions", store)

	ctx := context.Background()

	testAtom := &atom.Atom{Strings: map[string]string{"Token": "abc123"}}
	store.data["sess-1"] = testAtom

	got, err := s.Get(ctx, "kv://sessions/sess-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Strings["Token"] != "abc123" {
		t.Errorf("expected Token=abc123, got %v", got.Strings["Token"])
	}
}

func TestGet_Bucket(t *testing.T) {
	s := New()
	bucket := newMockBucket(documentSpec)
	_ = s.RegisterBucket("bcs://documents", bucket)

	ctx := context.Background()

	testAtom := &atom.Atom{Strings: map[string]string{"Title": "Report"}}
	bucket.data["report.pdf"] = &grub.AtomicObject{Data: testAtom}

	got, err := s.Get(ctx, "bcs://documents/report.pdf")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Strings["Title"] != "Report" {
		t.Errorf("expected Title=Report, got %v", got.Strings["Title"])
	}
}

func TestGet_NotFound(t *testing.T) {
	s := New()
	db := newMockDatabase("users", userSpec)
	_ = s.RegisterDatabase("db://users", db)

	ctx := context.Background()

	_, err := s.Get(ctx, "db://users/nonexistent")
	if !errors.Is(err, grub.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGet_ResourceNotFound(t *testing.T) {
	s := New()
	ctx := context.Background()

	_, err := s.Get(ctx, "db://nonexistent/123")
	if !errors.Is(err, ErrResourceNotFound) {
		t.Errorf("expected ErrResourceNotFound, got %v", err)
	}
}

func TestGet_KeyRequired(t *testing.T) {
	s := New()
	db := newMockDatabase("users", userSpec)
	_ = s.RegisterDatabase("db://users", db)

	ctx := context.Background()

	_, err := s.Get(ctx, "db://users")
	if !errors.Is(err, ErrKeyRequired) {
		t.Errorf("expected ErrKeyRequired, got %v", err)
	}
}

func TestSet_Database(t *testing.T) {
	s := New()
	db := newMockDatabase("users", userSpec)
	_ = s.RegisterDatabase("db://users", db)

	ctx := context.Background()
	testAtom := &atom.Atom{Strings: map[string]string{"Name": "Bob"}}

	err := s.Set(ctx, "db://users/456", testAtom)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Verify data was stored
	if db.data["456"].Strings["Name"] != "Bob" {
		t.Errorf("expected Name=Bob in stored data")
	}
}

func TestSet_Store(t *testing.T) {
	s := New()
	store := newMockStore(sessionSpec)
	_ = s.RegisterStore("kv://sessions", store)

	ctx := context.Background()
	testAtom := &atom.Atom{Strings: map[string]string{"Token": "xyz789"}}

	err := s.Set(ctx, "kv://sessions/sess-2", testAtom)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if store.data["sess-2"].Strings["Token"] != "xyz789" {
		t.Errorf("expected Token=xyz789 in stored data")
	}
}

func TestSet_BucketReturnsVariantMismatch(t *testing.T) {
	s := New()
	bucket := newMockBucket(documentSpec)
	_ = s.RegisterBucket("bcs://documents", bucket)

	ctx := context.Background()
	testAtom := &atom.Atom{}

	err := s.Set(ctx, "bcs://documents/file.txt", testAtom)
	if !errors.Is(err, ErrVariantMismatch) {
		t.Errorf("expected ErrVariantMismatch for bucket Set, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	s := New()
	db := newMockDatabase("users", userSpec)
	_ = s.RegisterDatabase("db://users", db)

	ctx := context.Background()
	db.data["123"] = &atom.Atom{}

	err := s.Delete(ctx, "db://users/123")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if _, ok := db.data["123"]; ok {
		t.Error("expected data to be deleted")
	}
}

func TestDelete_NotFound(t *testing.T) {
	s := New()
	db := newMockDatabase("users", userSpec)
	_ = s.RegisterDatabase("db://users", db)

	ctx := context.Background()

	err := s.Delete(ctx, "db://users/nonexistent")
	if !errors.Is(err, grub.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestExists(t *testing.T) {
	s := New()
	db := newMockDatabase("users", userSpec)
	_ = s.RegisterDatabase("db://users", db)

	ctx := context.Background()
	db.data["123"] = &atom.Atom{}

	exists, err := s.Exists(ctx, "db://users/123")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("expected exists=true")
	}

	exists, err = s.Exists(ctx, "db://users/nonexistent")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("expected exists=false")
	}
}

func TestQuery(t *testing.T) {
	s := New()
	db := newMockDatabase("users", userSpec)
	_ = s.RegisterDatabase("db://users", db)

	ctx := context.Background()
	db.data["1"] = &atom.Atom{Strings: map[string]string{"Name": "Alice"}}
	db.data["2"] = &atom.Atom{Strings: map[string]string{"Name": "Bob"}}

	results, err := s.Query(ctx, "db://users", grub.QueryAll, nil)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestQuery_VariantMismatch(t *testing.T) {
	s := New()
	store := newMockStore(sessionSpec)
	_ = s.RegisterStore("kv://sessions", store)

	ctx := context.Background()

	_, err := s.Query(ctx, "kv://sessions", grub.QueryAll, nil)
	if !errors.Is(err, ErrVariantMismatch) {
		t.Errorf("expected ErrVariantMismatch, got %v", err)
	}
}

func TestPut(t *testing.T) {
	s := New()
	bucket := newMockBucket(documentSpec)
	_ = s.RegisterBucket("bcs://documents", bucket)

	ctx := context.Background()
	obj := &grub.AtomicObject{
		Key:         "report.pdf",
		ContentType: "application/pdf",
		Data:        &atom.Atom{Strings: map[string]string{"Title": "Q4 Report"}},
	}

	err := s.Put(ctx, "bcs://documents/report.pdf", obj)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	if bucket.data["report.pdf"].ContentType != "application/pdf" {
		t.Errorf("expected ContentType=application/pdf")
	}
}

func TestPut_VariantMismatch(t *testing.T) {
	s := New()
	db := newMockDatabase("users", userSpec)
	_ = s.RegisterDatabase("db://users", db)

	ctx := context.Background()

	err := s.Put(ctx, "db://users/123", &grub.AtomicObject{})
	if !errors.Is(err, ErrVariantMismatch) {
		t.Errorf("expected ErrVariantMismatch, got %v", err)
	}
}
