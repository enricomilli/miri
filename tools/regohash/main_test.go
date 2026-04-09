package main

import (
	"os"
	"strings"
	"testing"
)

func TestHashBijection(t *testing.T) {
	seen := make(map[string]int)
	for i := 0; i < 1000; i++ {
		h := indexToHash(i)
		if len(h) != 4 {
			t.Errorf("hash %s is not 4 letters", h)
		}
		if _, ok := seen[h]; ok {
			t.Errorf("collision detected for index %d and %d: %s", i, seen[h], h)
		}
		seen[h] = i

		idx, err := hashToIndex(h)
		if err != nil {
			t.Errorf("error decoding %s: %v", h, err)
		}
		if idx != i {
			t.Errorf("expected %d, got %d", i, idx)
		}
	}
}

func TestReplaceCommand(t *testing.T) {
	// Setup temporary file
	tmpfile, err := os.CreateTemp("", "regohash_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := "line 1\nline 2\nline 3\nline 4\nline 5\n"
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// Replace lines from index 1 (line 2) to index 3 (line 4)
	startHash := indexToHash(1)
	endHash := indexToHash(3)

	newContent := "new line A\nnew line B\n"

	err = replaceLines(tmpfile.Name(), startHash, endHash, newContent)
	if err != nil {
		t.Fatalf("replaceLines failed: %v", err)
	}

	// Verify
	updated, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	expected := "line 1\nnew line A\nnew line B\nline 5\n"
	if string(updated) != expected {
		t.Errorf("expected %q, got %q", expected, string(updated))
	}
}

func TestHashIndexingAndEdgeCases(t *testing.T) {
	_, err := hashToIndex("aaa")
	if err == nil {
		t.Errorf("expected error for short hash")
	}
	_, err = hashToIndex("aaa1")
	if err == nil {
		t.Errorf("expected error for invalid characters")
	}
	_, err = hashToIndex("AAAA")
	if err == nil {
		t.Errorf("expected error for uppercase characters")
	}
}

func TestReplaceCommandInvalidBounds(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "regohash_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Write([]byte("line 1\nline 2\n"))
	tmpfile.Close()

	startHash := indexToHash(3) // Exceeds file length
	endHash := indexToHash(5)

	err = replaceLines(tmpfile.Name(), startHash, endHash, "new line\n")
	if err != nil {
		t.Fatalf("unexpected error for out of bounds replace: %v", err)
	}

	updated, _ := os.ReadFile(tmpfile.Name())
	expected := "line 1\nline 2\n\nnew line\n"
	if string(updated) != expected {
		t.Errorf("expected %q, got %q", expected, string(updated))
	}
}

func TestReplaceCommandStartAfterEnd(t *testing.T) {
	startHash := indexToHash(5)
	endHash := indexToHash(3)
	err := replaceLines("somefile", startHash, endHash, "new line\n")
	if err == nil || !strings.Contains(err.Error(), "start hash") {
		t.Errorf("expected error for start after end, got %v", err)
	}
}

func TestReplaceEmptyFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "regohash_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	startHash := indexToHash(0)
	endHash := indexToHash(0)

	err = replaceLines(tmpfile.Name(), startHash, endHash, "first line\n")
	if err != nil {
		t.Fatalf("unexpected error for empty file: %v", err)
	}

	updated, _ := os.ReadFile(tmpfile.Name())
	expected := "first line\n"
	if string(updated) != expected {
		t.Errorf("expected %q, got %q", expected, string(updated))
	}
}
