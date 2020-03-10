package skynet

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestWalkDirectory(t *testing.T) {
	testDir, err := ioutil.TempDir("", "testWalkDir")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(testDir)

	subDir, err := ioutil.TempDir(testDir, "subDir")
	if err != nil {
		t.Error(err)
	}

	// lexical order to match filepath.Walk
	testFiles := []string{
		filepath.Join(testDir, "one"),
		subDir,
		filepath.Join(subDir, "foo"),
		filepath.Join(testDir, "three"),
		filepath.Join(testDir, "two"),
	}

	for i, f := range testFiles {
		if i != 1 { // skip subDir which already exists
			if err = ioutil.WriteFile(f, make([]byte, 0), 0666); err != nil {
				t.Error(err)
			}
		}
	}

	files, err := walkDirectory(testDir)
	if err != nil {
		t.Error(err)
	}

	testFiles = append([]string{testDir}, testFiles...)
	if len(files) != len(testFiles) {
		t.Errorf("length %d of walked files != length %d of testFiles", len(files), len(testFiles))
	}
	for i, f := range files {
		if f != testFiles[i] {
			t.Errorf("file %s at index %d != testFile %s at same index", f, i, testFiles[i])
		}
	}
}
