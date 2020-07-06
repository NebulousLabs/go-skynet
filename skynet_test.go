package skynet

import (
	"testing"
)

// TestWalkDirectory tests directory walking.
func TestWalkDirectory(t *testing.T) {
	const testDir = "testdata"

	files, err := walkDirectory(testDir)
	if err != nil {
		t.Error(err)
	}
	expectedFiles := []string{
		"testdata/dir1/file3.txt",
		"testdata/file1.txt",
		"testdata/file2.txt",
	}

	if len(files) != len(expectedFiles) {
		t.Errorf("expected %v files, got %v", len(expectedFiles), len(files))
	}
	for i, f := range files {
		if f != expectedFiles[i] {
			t.Errorf("file %s at index %d != expected file %s at same index", f, i, expectedFiles[i])
		}
	}
}
