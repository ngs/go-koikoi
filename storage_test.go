package main

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestStorageRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "data.json")

	if _, err := readFileData(path); err == nil {
		t.Fatal("readFileData() on missing file should return error")
	}

	want := []byte(`{"key":"value"}`)
	if err := writeFileData(path, want); err != nil {
		t.Fatalf("writeFileData() error = %v", err)
	}

	got, err := readFileData(path)
	if err != nil {
		t.Fatalf("readFileData() error = %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("readFileData() = %q, want %q", got, want)
	}

	removeFileData(path)
	if _, err := readFileData(path); err == nil {
		t.Fatal("readFileData() after removeFileData() should return error")
	}

	// 存在しないファイルの削除は no-op
	removeFileData(path)
}

func TestWriteFileDataOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")

	if err := writeFileData(path, []byte("first")); err != nil {
		t.Fatalf("writeFileData() error = %v", err)
	}
	if err := writeFileData(path, []byte("second")); err != nil {
		t.Fatalf("writeFileData() overwrite error = %v", err)
	}

	got, err := readFileData(path)
	if err != nil {
		t.Fatalf("readFileData() error = %v", err)
	}
	if string(got) != "second" {
		t.Errorf("readFileData() = %q, want %q", got, "second")
	}
}
