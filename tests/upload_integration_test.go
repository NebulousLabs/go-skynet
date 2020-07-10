package tests

import (
	"path/filepath"
	"strings"
	"testing"

	skynet "github.com/NebulousLabs/go-skynet"
	"gopkg.in/h2non/gock.v1"
)

// TestUploadFile tests uploading a single file.
func TestUploadFile(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.Observe(interceptRequest)

	const srcFile = "../testdata/file1.txt"
	const skylink = "testskynet"
	const sialink = skynet.URISkynetPrefix + skylink

	// Test that uploading a nonexistent file fails.

	_, err := skynet.UploadFile("this-should-not-exist.txt", skynet.DefaultUploadOptions)
	if !strings.Contains(err.Error(), "no such file or directory") {
		t.Fatalf("expected ErrNotExist error, got %v", err)
	}

	// Test uploading a file.

	// Upload file request.
	opts := skynet.DefaultUploadOptions
	gock.New(skynet.DefaultPortalURL).
		Post(opts.PortalUploadPath).
		Reply(200).
		JSON(map[string]string{"skylink": skylink})

	sialink2, err := skynet.UploadFile(srcFile, opts)
	if err != nil {
		t.Fatal(err)
	}
	if sialink2 != sialink {
		t.Fatalf("expected sialink %v, got %v", sialink, sialink2)
	}

	// Test uploading a file with a custom filename.

	interceptedRequest = ""

	opts.CustomFilename = "foobar"
	gock.New(skynet.DefaultPortalURL).
		Post(opts.PortalUploadPath).
		Reply(200).
		JSON(map[string]string{"skylink": skylink})

	sialink2, err = skynet.UploadFile(srcFile, opts)
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

// TestUploadDirectory tests uploading an entire directory.
func TestUploadDirectory(t *testing.T) {
	defer gock.Off()
	gock.Observe(interceptRequest)

	const srcDir = "../testdata"
	const skylink = "testskynet"
	const sialink = skynet.URISkynetPrefix + skylink

	filename := filepath.Base(srcDir)

	// Upload a directory.

	opts := skynet.DefaultUploadOptions
	gock.New(skynet.DefaultPortalURL).
		Post(opts.PortalUploadPath).
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
	if count != 3 {
		t.Fatalf("expected %v files sent, got %v", 3, count)
	}

	if sialink2 != sialink {
		t.Fatalf("expected sialink %v, got %v", sialink, sialink2)
	}

	// Upload a directory with a custom dirname.

	opts = skynet.DefaultUploadOptions
	opts.CustomDirname = "barfoo"
	gock.New(skynet.DefaultPortalURL).
		Post(opts.PortalUploadPath).
		MatchParam("filename", "barfoo").
		Reply(200).
		JSON(map[string]string{"skylink": skylink})

	interceptedRequest = ""

	sialink2, err = skynet.UploadDirectory(srcDir, opts)
	if err != nil {
		t.Fatal(err)
	}

	count = strings.Count(interceptedRequest, "Content-Disposition")
	if count != 3 {
		t.Fatalf("expected %v files sent, got %v", 3, count)
	}

	if sialink2 != sialink {
		t.Fatalf("expected sialink %v, got %v", sialink, sialink2)
	}

	// Verify we don't have pending mocks.
	if !gock.IsDone() {
		t.Fatal("test finished with pending mocks")
	}
}
