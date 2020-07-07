package tests

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	skynet "github.com/NebulousLabs/go-skynet"
	"gopkg.in/h2non/gock.v1"
)

var (
	// interceptRequest is a gock observer function that intercepts requests and
	// writes them to `interceptedRequest`.
	interceptRequest gock.ObserverFunc = func(request *http.Request, mock gock.Mock) {
		bytes, _ := httputil.DumpRequestOut(request, true)
		interceptedRequest = string(bytes)
	}

	// interceptedRequest contains the raw data of intercepted requests.
	interceptedRequest string
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
	if !os.IsNotExist(err) {
		t.Fatalf("expected IsNotExist error, got %v", err)
	}

	// Test uploading a file.

	// Upload file request.
	gock.New(skynet.DefaultUploadOptions.PortalURL).
		Post(skynet.DefaultUploadOptions.PortalUploadPath).
		Reply(200).
		JSON(map[string]string{"skylink": skylink})

	sialink2, err := skynet.UploadFile(srcFile, skynet.DefaultUploadOptions)
	if err != nil {
		t.Fatal(err)
	}
	if sialink2 != sialink {
		t.Fatalf("expected sialink %v, got %v", sialink, sialink2)
	}

	// Test uploading a file with a custom filename.

	interceptedRequest = ""

	gock.New(skynet.DefaultUploadOptions.PortalURL).
		Post(skynet.DefaultUploadOptions.PortalUploadPath).
		Reply(200).
		JSON(map[string]string{"skylink": skylink})

	opts := skynet.DefaultUploadOptions
	opts.CustomFilename = "foobar"
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

// TestDownloadFile tests downloading a single file.
func TestDownloadFile(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution

	const srcFile = "../testdata/file1.txt"
	const skylink = "testskynet"
	const sialink = skynet.URISkynetPrefix + skylink

	file, err := ioutil.TempFile("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	dstFile := file.Name()

	// Download file request.
	//
	// Match against the full URL, including the skylink.
	urlpath := strings.TrimRight(skynet.DefaultDownloadOptions.PortalDownloadPath, "/") + "/" + skylink
	gock.New(skynet.DefaultDownloadOptions.PortalURL).
		Get(urlpath).
		Reply(200).
		BodyString("test\n")

	// Pass the full sialink to verify that the prefix is trimmed.
	err = skynet.DownloadFile(dstFile, sialink, skynet.DefaultDownloadOptions)
	if err != nil {
		t.Fatal(err)
	}

	// Check file equality.
	f1, err1 := ioutil.ReadFile(srcFile)
	if err1 != nil {
		t.Fatal(err1)
	}
	f2, err2 := ioutil.ReadFile(path.Clean(dstFile))
	if err2 != nil {
		t.Fatal(err2)
	}
	if !bytes.Equal(f1, f2) {
		t.Fatalf("Downloaded file at %v did not equal uploaded file %v", dstFile, srcFile)
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

	gock.New(skynet.DefaultUploadOptions.PortalURL).
		Post(skynet.DefaultUploadOptions.PortalUploadPath).
		MatchParam("filename", filename).
		Reply(200).
		JSON(map[string]string{"skylink": skylink})

	interceptedRequest = ""

	sialink2, err := skynet.UploadDirectory(srcDir, skynet.DefaultUploadOptions)
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

	gock.New(skynet.DefaultUploadOptions.PortalURL).
		Post(skynet.DefaultUploadOptions.PortalUploadPath).
		MatchParam("filename", "barfoo").
		Reply(200).
		JSON(map[string]string{"skylink": skylink})

	interceptedRequest = ""

	opts := skynet.DefaultUploadOptions
	opts.CustomDirname = "barfoo"

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
