package tests

import (
	"bytes"
	"io/ioutil"
	"path"
	"path/filepath"
	"testing"

	skynet "github.com/NebulousLabs/go-skynet"
	"gopkg.in/h2non/gock.v1"
)

// TestUploadAndDownloadFile tests uploading and downloading a single file.
func TestUploadAndDownloadFile(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution

	const srcFile = "../testdata/file1.txt"
	const skylink = skynet.URISkynetPrefix + "testskynet"

	// Test upload.

	gock.New(skynet.DefaultUploadOptions.PortalURL).
		Post(skynet.DefaultUploadOptions.PortalUploadPath).
		Reply(200).
		JSON(map[string]string{"skylink": "testskynet"})

	skylink2, err := skynet.UploadFile(srcFile, skynet.DefaultUploadOptions)
	if err != nil {
		t.Fatal(err)
	}
	if skylink2 != skylink {
		t.Fatalf("expected skylink %v, got %v", skylink, skylink2)
	}

	// Test download.

	gock.New(skynet.DefaultDownloadOptions.PortalURL).
		Get(skynet.DefaultDownloadOptions.PortalDownloadPath).
		Reply(200).
		BodyString("test\n")

	file, err := ioutil.TempFile("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	dstFile := file.Name()

	err = skynet.DownloadFile(dstFile, skylink, skynet.DefaultDownloadOptions)
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
}

// TestUploadDirectory tests uploading an entire directory.
func TestUploadDirectory(t *testing.T) {
	defer gock.Off()

	const srcDir = "../testdata"
	const skylink = skynet.URISkynetPrefix + "testskynet"

	filename := filepath.Base(srcDir)

	gock.New(skynet.DefaultUploadOptions.PortalURL).
		Post(skynet.DefaultUploadOptions.PortalUploadPath).
		MatchParam("filename", filename).
		Reply(200).
		JSON(map[string]string{"skylink": "testskynet"})

	skylink2, err := skynet.UploadDirectory(srcDir, skynet.DefaultUploadOptions)
	if err != nil {
		t.Fatal(err)
	}
	if skylink2 != skylink {
		t.Fatalf("expected skylink %v, got %v", skylink, skylink2)
	}
}
