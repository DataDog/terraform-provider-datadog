package codegen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFormatGoSource_Valid(t *testing.T) {
	src := []byte("package main\n\nfunc main() { x := 1\n_ = x}\n")
	formatted, err := FormatGoSource(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "package main\n\nfunc main() {\n\tx := 1\n\t_ = x\n}\n"
	if string(formatted) != expected {
		t.Errorf("formatted = %q, want %q", string(formatted), expected)
	}
}

func TestFormatGoSource_Invalid(t *testing.T) {
	src := []byte("this is not valid go code {{{")
	_, err := FormatGoSource(src)
	if err == nil {
		t.Fatal("expected error for invalid Go source")
	}
}

func TestWriteGoFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")
	content := []byte("package test\n\nfunc Hello() string { return \"hello\" }\n")

	if err := WriteGoFile(path, content); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty file")
	}

	// File should be properly formatted
	if string(data) != "package test\n\nfunc Hello() string { return \"hello\" }\n" {
		t.Errorf("file content = %q", string(data))
	}
}

func TestWriteGoFile_CreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "test.go")
	content := []byte("package test\n")

	if err := WriteGoFile(path, content); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected file to exist")
	}
}

func TestWriteIfNotExists_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.go")
	content := []byte("package test\n")

	if err := WriteIfNotExists(path, content); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected file to be created")
	}
}

func TestWriteIfNotExists_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "existing.go")

	original := []byte("package original\n")
	if err := os.WriteFile(path, original, 0644); err != nil {
		t.Fatalf("writing original: %v", err)
	}

	// Try to write different content
	newContent := []byte("package replaced\n")
	if err := WriteIfNotExists(path, newContent); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// File should still have original content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	if string(data) != string(original) {
		t.Errorf("file was overwritten: got %q, want %q", string(data), string(original))
	}
}
