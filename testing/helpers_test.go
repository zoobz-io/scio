package testing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zoobz-io/atom"
	"github.com/zoobz-io/grub"
)

func TestWithTimeout(t *testing.T) {
	ctx := WithTimeout(t, 100*time.Millisecond)
	if ctx == nil {
		t.Fatal("expected context, got nil")
	}

	select {
	case <-ctx.Done():
		t.Fatal("context should not be done yet")
	default:
	}
}

func TestAssertNoError(t *testing.T) {
	// Should not panic
	AssertNoError(t, nil)
}

func TestAssertEqual(t *testing.T) {
	AssertEqual(t, 42, 42)
	AssertEqual(t, "hello", "hello")
}

func TestAssertTrue(t *testing.T) {
	AssertTrue(t, true)
}

func TestAssertFalse(t *testing.T) {
	AssertFalse(t, false)
}

func TestMockDatabase(t *testing.T) {
	spec := TestSpec("User", "ID", "Name")
	db := NewMockDatabase("users", spec)

	if db.Table() != "users" {
		t.Errorf("expected table 'users', got %q", db.Table())
	}
	if db.Spec().FQDN != spec.FQDN {
		t.Errorf("expected FQDN %q, got %q", spec.FQDN, db.Spec().FQDN)
	}

	ctx := context.Background()
	testAtom := &atom.Atom{Strings: map[string]string{"Name": "Alice"}}

	// Set and Get
	err := db.Set(ctx, "1", testAtom)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, err := db.Get(ctx, "1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Strings["Name"] != "Alice" {
		t.Errorf("expected Name=Alice, got %v", got.Strings["Name"])
	}

	// Exists
	exists, _ := db.Exists(ctx, "1")
	if !exists {
		t.Error("expected exists=true")
	}

	// Delete
	err = db.Delete(ctx, "1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = db.Get(ctx, "1")
	if !errors.Is(err, grub.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMockStore(t *testing.T) {
	spec := TestSpec("Session", "Token")
	store := NewMockStore(spec)

	ctx := context.Background()
	testAtom := &atom.Atom{Strings: map[string]string{"Token": "abc"}}

	err := store.Set(ctx, "sess-1", testAtom, 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, err := store.Get(ctx, "sess-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Strings["Token"] != "abc" {
		t.Errorf("expected Token=abc, got %v", got.Strings["Token"])
	}
}

func TestMockBucket(t *testing.T) {
	spec := TestSpec("Document", "Title")
	bucket := NewMockBucket(spec)

	ctx := context.Background()
	obj := &grub.AtomicObject{
		Key:  "doc.pdf",
		Data: &atom.Atom{Strings: map[string]string{"Title": "Report"}},
	}

	err := bucket.Put(ctx, "doc.pdf", obj)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	got, err := bucket.Get(ctx, "doc.pdf")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Data.Strings["Title"] != "Report" {
		t.Errorf("expected Title=Report, got %v", got.Data.Strings["Title"])
	}
}

func TestTestSpec(t *testing.T) {
	spec := TestSpec("User", "ID", "Email", "Name")

	if spec.FQDN != "github.com/test.User" {
		t.Errorf("expected FQDN github.com/test.User, got %q", spec.FQDN)
	}
	if len(spec.Fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(spec.Fields))
	}
	if spec.Fields[0].Name != "ID" {
		t.Errorf("expected first field ID, got %q", spec.Fields[0].Name)
	}
}
