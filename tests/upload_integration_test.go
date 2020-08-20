package tests

import (
	"path/filepath"
	"strings"
	"testing"

	skynet "github.com/NebulousLabs/go-skynet"
	"gopkg.in/h2non/gock.v1"
)

const (
	srcDir     = "../testdata"
	srcFile    = "../testdata/file1.txt"
	skylink    = "XABvi7JtJbQSMAcDwnUnmp2FKDPjg8_tTTFP4BwMSxVdEg"
	skykeyName = "testcreateskykey"
	skykeyID   = "pJAPPfWkWXpss3BvMDCJCw=="
)

var (
	sialink = skynet.URISkynetPrefix + skylink
)

// TestUploadFile tests uploading a single file.
func TestUploadFile(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.Observe(interceptRequest)

	interceptedRequest = ""

	// Test that uploading a nonexistent file fails.

	_, err := skynet.UploadFile("this-should-not-exist.txt", skynet.DefaultUploadOptions)
	if !strings.Contains(err.Error(), "no such file or directory") {
		t.Fatalf("expected ErrNotExist error, got %v", err)
	}

	// Test uploading a file.

	// Upload file request.
	opts := skynet.DefaultUploadOptions
	gock.New(skynet.DefaultPortalURL).
		Post(opts.EndpointPath).
		Reply(200).
		JSON(map[string]string{"skylink": skylink})

	sialink2, err := skynet.UploadFile(srcFile, opts)
	if err != nil {
		t.Fatal(err)
	}
	if sialink2 != sialink {
		t.Fatalf("expected sialink %v, got %v", sialink, sialink2)
	}

	// Check that the content type is set.
	if !strings.Contains(interceptedRequest, "Content-Type: multipart/form-data;") {
		t.Fatal("Content-Type header incorrect")
	}

	// Verify we don't have pending mocks.
	if !gock.IsDone() {
		t.Fatal("test finished with pending mocks")
	}
}

// TestUploadFileWithAPIKey tests uploading a single file with authentication.
func TestUploadFileWithAPIKey(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution

	// Test uploading a file.

	// Upload file request.
	opts := skynet.DefaultUploadOptions
	opts.APIKey = "foobar"
	gock.New(skynet.DefaultPortalURL).
		Post(opts.EndpointPath).
		MatchHeader("Authorization", "Basic OmZvb2Jhcg==").
		Reply(200).
		JSON(map[string]string{"skylink": skylink})

	sialink2, err := skynet.UploadFile(srcFile, opts)
	if err != nil {
		t.Fatal(err)
	}
	if sialink2 != sialink {
		t.Fatalf("expected sialink %v, got %v", sialink, sialink2)
	}

	// Verify we don't have pending mocks.
	if !gock.IsDone() {
		t.Fatal("test finished with pending mocks")
	}
}

// TestUploadFileCustomName tests uploading a single file with a custom
// filename.
func TestUploadFileCustomName(t *testing.T) {
	defer gock.Off()
	gock.Observe(interceptRequest)

	// Test uploading a file with a custom filename.

	interceptedRequest = ""

	opts := skynet.DefaultUploadOptions
	opts.CustomFilename = "foobar"
	gock.New(skynet.DefaultPortalURL).
		Post(opts.EndpointPath).
		Reply(200).
		JSON(map[string]string{"skylink": skylink})

	sialink2, err := skynet.UploadFile(srcFile, opts)
	if err != nil {
		t.Fatal(err)
	}
	if sialink2 != sialink {
		t.Fatalf("expected sialink %v, got %v", sialink, sialink2)
	}

	if !strings.Contains(interceptedRequest, "Content-Disposition: form-data; name=\"file\"; filename=\"foobar\"") {
		t.Fatal("expected request body to contain foobar")
	}
	count := strings.Count(interceptedRequest, "Content-Disposition")
	if count != 1 {
		t.Fatalf("expected %v files sent, got %v", 1, count)
	}

	// Verify we don't have pending mocks.
	if !gock.IsDone() {
		t.Fatal("test finished with pending mocks")
	}
}

// TestUploadFileSkykey tests uploading a file with either a skykey name or
// skykey id set.
func TestUploadFileSkykey(t *testing.T) {
	defer gock.Off()

	// Test uploading a file with a skykey name set.

	opts := skynet.DefaultUploadOptions
	opts.SkykeyName = skykeyName
	gock.New(skynet.DefaultPortalURL).
		Post(opts.EndpointPath).
		MatchParam("skykeyname", skykeyName).
		Reply(200).
		JSON(map[string]string{"skylink": skylink})

	sialink2, err := skynet.UploadFile(srcFile, opts)
	if err != nil {
		t.Fatal(err)
	}
	if sialink2 != sialink {
		t.Fatalf("expected sialink %v, got %v", sialink, sialink2)
	}

	// Test uploading a file with a skykey id set.

	opts = skynet.DefaultUploadOptions
	opts.SkykeyID = skykeyID
	gock.New(skynet.DefaultPortalURL).
		Post(opts.EndpointPath).
		MatchParam("skykeyid", skykeyID).
		Reply(200).
		JSON(map[string]string{"skylink": skylink})

	sialink2, err = skynet.UploadFile(srcFile, opts)
	if err != nil {
		t.Fatal(err)
	}
	if sialink2 != sialink {
		t.Fatalf("expected sialink %v, got %v", sialink, sialink2)
	}

	// Verify we don't have pending mocks.
	if !gock.IsDone() {
		t.Fatal("test finished with pending mocks")
	}
}

// TestUploadDirectory tests uploading an entire directory.
func TestUploadDirectory(t *testing.T) {
	defer gock.Off()
	gock.Observe(interceptRequest)

	filename := filepath.Base(srcDir)

	// Upload a directory.

	opts := skynet.DefaultUploadOptions
	gock.New(skynet.DefaultPortalURL).
		Post(opts.EndpointPath).
		MatchParam("filename", filename).
		Reply(200).
		JSON(map[string]string{"skylink": skylink})

	interceptedRequest = ""

	sialink2, err := skynet.UploadDirectory(srcDir, opts)
	if err != nil {
		t.Fatal(err)
	}

	// Check that the request contained the files in `testdata/`.
	if !strings.Contains(interceptedRequest, "Content-Disposition: form-data; name=\"files[]\"; filename=\"file1.txt\"") {
		t.Fatal("expected request body to contain file1.txt")
	}
	if !strings.Contains(interceptedRequest, "Content-Disposition: form-data; name=\"files[]\"; filename=\"file2.txt\"") {
		t.Fatal("expected request body to contain file2.txt")
	}
	if !strings.Contains(interceptedRequest, "Content-Disposition: form-data; name=\"files[]\"; filename=\"dir1/file3.txt\"") {
		t.Fatal("expected request body to contain dir1/file3.txt")
	}
	// The request should not contain a nonexistent file.
	if strings.Contains(interceptedRequest, "Content-Disposition: form-data; name=\"files[]\"; filename=\"file0.txt\"") {
		t.Fatal("did not expect request body to contain file0.txt")
	}
	count := strings.Count(interceptedRequest, "Content-Disposition")
	if count != 4 {
		t.Fatalf("expected %v files sent, got %v", 4, count)
	}

	if sialink2 != sialink {
		t.Fatalf("expected sialink %v, got %v", sialink, sialink2)
	}

	// Check that the content type is set.
	if !strings.Contains(interceptedRequest, "Content-Type: multipart/form-data;") {
		t.Fatal("Content-Type header incorrect")
	}

	// Verify we don't have pending mocks.
	if !gock.IsDone() {
		t.Fatal("test finished with pending mocks")
	}
}

// TestUploadDirectoryContentTypes tests that uploading a directory sets the
// correct content types for subfiles.
func TestUploadDirectoryContentTypes(t *testing.T) {
	defer gock.Off()
	gock.Observe(interceptRequest)

	// Upload a directory.

	opts := skynet.DefaultUploadOptions
	gock.New(skynet.DefaultPortalURL).
		Post(opts.EndpointPath).
		Reply(200).
		JSON(map[string]string{"skylink": skylink})

	interceptedRequest = ""

	_, err := skynet.UploadDirectory(srcDir, opts)
	if err != nil {
		t.Fatal(err)
	}

	print(interceptedRequest)
	expectedHeader := "Content-Disposition: form-data; name=\"files[]\"; filename=\"index.html\"\r\nContent-Type: text/html; charset=utf-8"
	if !strings.Contains(interceptedRequest, expectedHeader) {
		t.Fatal("did not find expected header")
	}

	// Verify we don't have pending mocks.
	if !gock.IsDone() {
		t.Fatal("test finished with pending mocks")
	}
}

// TestUploadDirectoryCustomName tests uploading a directory with a custom name.
func TestUploadDirectoryCustomName(t *testing.T) {
	defer gock.Off()
	gock.Observe(interceptRequest)

	// Upload a directory with a custom dirname.

	opts := skynet.DefaultUploadOptions
	opts.CustomDirname = "barfoo"
	gock.New(skynet.DefaultPortalURL).
		Post(opts.EndpointPath).
		MatchParam("filename", "barfoo").
		Reply(200).
		JSON(map[string]string{"skylink": skylink})

	interceptedRequest = ""

	_, err := skynet.UploadDirectory(srcDir, opts)
	if err != nil {
		t.Fatal(err)
	}

	count := strings.Count(interceptedRequest, "Content-Disposition")
	if count != 4 {
		t.Fatalf("expected %v files sent, got %v", 4, count)
	}

	// Verify we don't have pending mocks.
	if !gock.IsDone() {
		t.Fatal("test finished with pending mocks")
	}
}

// TestUploadDirectorySkykey tests uploading a directory with either a skykey
// name or skykey id set.
func TestUploadDirectorySkykey(t *testing.T) {
	defer gock.Off()

	filename := filepath.Base(srcDir)

	// Upload a directory with a skykey name set.

	opts := skynet.DefaultUploadOptions
	opts.SkykeyName = skykeyName
	gock.New(skynet.DefaultPortalURL).
		Post(opts.EndpointPath).
		MatchParam("filename", filename).
		MatchParam("skykeyname", skykeyName).
		Reply(200).
		JSON(map[string]string{"skylink": skylink})

	interceptedRequest = ""

	sialink2, err := skynet.UploadDirectory(srcDir, opts)
	if err != nil {
		t.Fatal(err)
	}

	count := strings.Count(interceptedRequest, "Content-Disposition")
	if count != 4 {
		t.Fatalf("expected %v files sent, got %v", 4, count)
	}

	if sialink2 != sialink {
		t.Fatalf("expected sialink %v, got %v", sialink, sialink2)
	}

	// Upload a directory with a skykey ID set.

	opts = skynet.DefaultUploadOptions
	opts.SkykeyID = skykeyID
	gock.New(skynet.DefaultPortalURL).
		Post(opts.EndpointPath).
		MatchParam("filename", filename).
		MatchParam("skykeyid", skykeyID).
		Reply(200).
		JSON(map[string]string{"skylink": skylink})

	interceptedRequest = ""

	sialink2, err = skynet.UploadDirectory(srcDir, opts)
	if err != nil {
		t.Fatal(err)
	}

	count = strings.Count(interceptedRequest, "Content-Disposition")
	if count != 4 {
		t.Fatalf("expected %v files sent, got %v", 4, count)
	}

	if sialink2 != sialink {
		t.Fatalf("expected sialink %v, got %v", sialink, sialink2)
	}

	// Verify we don't have pending mocks.
	if !gock.IsDone() {
		t.Fatal("test finished with pending mocks")
	}
}
