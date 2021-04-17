package skynet

import (
	"net/url"
	"testing"
)

var (
	portalURL = DefaultPortalURL()
)

// TestMakeURL tests making URLs.
func TestMakeURL(t *testing.T) {
	values1 := url.Values{}
	values2 := url.Values{}
	values2.Set("foo", "bar")
	values2.Set("bar", "foo")
	tests := []struct {
		url, path, extraPath string
		values               url.Values
		out                  string
	}{
		{
			portalURL, "test", "",
			nil,
			portalURL + "/test",
		},
		{
			portalURL, "test", "skylink",
			values1,
			portalURL + "/test/skylink",
		},
		{
			portalURL, "/", "",
			nil,
			portalURL + "/",
		},
		{
			portalURL, "/", "skylink",
			nil,
			portalURL + "/skylink",
		},
		{
			portalURL + "/test", "skylink", "",
			nil,
			portalURL + "/test/skylink",
		},
		{
			portalURL, "skynet/skyfile", "",
			values2,
			portalURL + "/skynet/skyfile?bar=foo&foo=bar",
		},
		{
			portalURL, "//test/", "",
			values2,
			portalURL + "/test/?bar=foo&foo=bar",
		},
	}

	for _, test := range tests {
		url := makeURL(test.url, test.path, test.extraPath, test.values)
		if url != test.out {
			t.Fatalf("expected %v, got %v", test.out, url)
		}
	}
}

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
		"testdata/index.html",
		"testdata/indexhtml",
		"testdata/jsonFile1.json",
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
