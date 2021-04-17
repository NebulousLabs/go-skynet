package tests

import (
	"bytes"
	"github.com/NebulousLabs/go-skynet/v2"
	"gopkg.in/h2non/gock.v1"
	"io/ioutil"
	"strings"
	"testing"
)

// TestGetJSON tests get of JSON from registry.
func TestGetJSON(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution

	const srcFile = "../testdata/jsonFile1.json"
	const skylink = "AADeqJt8vPZtW9Nm_Hc5C5EKPmZhYUZGeBqvibofIMMHtg"

	registryFile, err := ioutil.ReadFile(srcFile)
	if err != nil {
		t.Fatal(err)
	}

	opts := skynet.DefaultDownloadOptions
	urlpath := strings.TrimRight(opts.EndpointPath, "/") + "/" + skylink
	gock.New(skynet.DefaultPortalURL()).
		Get(urlpath).
		Reply(200).
		BodyString(string(registryFile))

	gock.New(skynet.DefaultPortalURL()).
		Get(skynet.RegistryEndpoint).
		Reply(200).
		BodyString(`{
			"data":"41414465714a743876505a7457394e6d5f4863354335454b506d5a6859555a476542717669626f66494d4d487467",
			"revision": 2,
			"signature": "f3dc30c2255254a7ffd64e767e15f8b9dc908491907c79afb3a1b24ee3b9602f10ff01bce22e1e700f502190fff4ee209f5b32e4c2b9e1ef6b0bed0c2b558406"
		}`)

	jsonReader, err := client.GetJSON(
		"4a964fa1cb329d066aedcf7fc03a249eeea3cf2461811090b287daaaec37ab36",
		"TEST_KEY",
	)
	if err != nil {
		t.Fatal(err)
	}

	json, err := ioutil.ReadAll(jsonReader)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(registryFile, json) {
		t.Fatalf("registryFile and fetched JSON did not equal")
	}

	// Verify we don't have pending mocks.
	if !gock.IsDone() {
		t.Fatal("test finished with pending mocks")
	}
}

// TestGetJSON with an invalid signature.
func TestGetJSON_invalid_signature(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution

	const srcFile = "../testdata/jsonFile1.json"
	const skylink = "AADeqJt8vPZtW9Nm_Hc5C5EKPmZhYUZGeBqvibofIMMHtg"

	registryFile, err := ioutil.ReadFile(srcFile)
	if err != nil {
		t.Fatal(err)
	}

	opts := skynet.DefaultDownloadOptions
	urlpath := strings.TrimRight(opts.EndpointPath, "/") + "/" + skylink
	gock.New(skynet.DefaultPortalURL()).
		Get(urlpath).
		Reply(200).
		BodyString(string(registryFile))

	gock.New(skynet.DefaultPortalURL()).
		Get(skynet.RegistryEndpoint).
		Reply(200).
		BodyString(`{
			"data":"41414465714a743876505a7457394e6d5f4863354335454b506d5a6859555a476542717669626f66494d4d487467",
			"revision": 2,
			"signature": "f3dc30c2255254a7ffd64e767e15f8b9dc908491907c79afb3a1b24e"
		}`)

	_, err = client.GetJSON(
		"4a964fa1cb329d066aedcf7fc03a249eeea3cf2461811090b287daaaec37ab36",
		"TEST_KEY",
	)
	if err == nil {
		t.Fatal("signature should be invalid")
	}
}

// TestGetJSON with an invalid skylink.
func TestGetJSON_invalid_skylink(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution

	const srcFile = "../testdata/jsonFile1.json"
	const skylink = "invalid_skylink"

	registryFile, err := ioutil.ReadFile(srcFile)
	if err != nil {
		t.Fatal(err)
	}

	opts := skynet.DefaultDownloadOptions
	urlpath := strings.TrimRight(opts.EndpointPath, "/") + "/" + skylink
	gock.New(skynet.DefaultPortalURL()).
		Get(urlpath).
		Reply(200).
		BodyString(string(registryFile))

	gock.New(skynet.DefaultPortalURL()).
		Get(skynet.RegistryEndpoint).
		Reply(200).
		BodyString(`{
			"data":"41414465714a743876505a7457394e6d5f4863354335454b506d5a6859555a476542717669626f66494d4d487467",
			"revision": 2,
			"signature": "f3dc30c2255254a7ffd64e767e15f8b9dc908491907c79afb3a1b24ee3b9602f10ff01bce22e1e700f502190fff4ee209f5b32e4c2b9e1ef6b0bed0c2b558406"
		}`)

	_, err = client.GetJSON(
		"4a964fa1cb329d066aedcf7fc03a249eeea3cf2461811090b287daaaec37ab36",
		"TEST_KEY",
	)
	if err == nil {
		t.Fatal("skylink should be invalid")
	}
}

// TestSetJSON tests setting of new JSON into the registry.
func TestSetJSON(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.Observe(interceptRequest)

	interceptedRequest = ""

	const skylink = "AADeqJt8vPZtW9Nm_Hc5C5EKPmZhYUZGeBqvibofIMMHtg"

	opts := skynet.DefaultUploadOptions
	gock.New(skynet.DefaultPortalURL()).
		Post(opts.EndpointPath).
		Reply(200).
		JSON(map[string]string{"skylink": skylink})

	gock.New(skynet.DefaultPortalURL()).
		Post(skynet.RegistryEndpoint).
		Reply(204)

	var newRevision uint64 = 10

	err := client.SetJSON(
		"191907ba9a408d2851cd0fa68c2f264ec8924d0d6ff64fc858070ee3b77b25524a964fa1cb329d066aedcf7fc03a249eeea3cf2461811090b287daaaec37ab36",
		"TEST_KEY",
		bytes.NewBuffer([]byte("test2")),
		&newRevision,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Check that the content type is set.
	if !strings.Contains(interceptedRequest, `"revision":10`) {
		t.Fatal("incorrect revision bump")
	}

	// Verify we don't have pending mocks.
	if !gock.IsDone() {
		t.Fatal("test finished with pending mocks")
	}
}

// TestSetJSON tests setting of new JSON into the registry with no revision provided.
func TestSetJSON_no_revision_provided(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.Observe(interceptRequest)

	interceptedRequest = ""

	const skylink = "AADeqJt8vPZtW9Nm_Hc5C5EKPmZhYUZGeBqvibofIMMHtg"

	gock.New(skynet.DefaultPortalURL()).
		Get(skynet.RegistryEndpoint).
		Reply(200).
		BodyString(`{
			"data":"41414465714a743876505a7457394e6d5f4863354335454b506d5a6859555a476542717669626f66494d4d487467",
			"revision": 2,
			"signature": "f3dc30c2255254a7ffd64e767e15f8b9dc908491907c79afb3a1b24ee3b9602f10ff01bce22e1e700f502190fff4ee209f5b32e4c2b9e1ef6b0bed0c2b558406"
		}`)

	opts := skynet.DefaultUploadOptions
	gock.New(skynet.DefaultPortalURL()).
		Post(opts.EndpointPath).
		Reply(200).
		JSON(map[string]string{"skylink": skylink})

	gock.New(skynet.DefaultPortalURL()).
		Post(skynet.RegistryEndpoint).
		Reply(204).
		BodyString("test")

	err := client.SetJSON(
		"191907ba9a408d2851cd0fa68c2f264ec8924d0d6ff64fc858070ee3b77b25524a964fa1cb329d066aedcf7fc03a249eeea3cf2461811090b287daaaec37ab36",
		"TEST_KEY",
		bytes.NewBuffer([]byte("test2")),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Check that the content type is set.
	if !strings.Contains(interceptedRequest, `"revision":3`) {
		t.Fatal("incorrect revision bump")
	}

	// Verify we don't have pending mocks.
	if !gock.IsDone() {
		t.Fatal("test finished with pending mocks")
	}
}

// TestSetJSON tests setting of new JSON into the registry with no revision provided.
func TestSetJSON_no_revision_provided_server_error(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.Observe(interceptRequest)

	interceptedRequest = ""

	const skylink = "AADeqJt8vPZtW9Nm_Hc5C5EKPmZhYUZGeBqvibofIMMHtg"

	gock.New(skynet.DefaultPortalURL()).
		Get(skynet.RegistryEndpoint).
		Reply(200).
		BodyString(`{
			"data":"41414465714a743876505a7457394e6d5f4863354335454b506d5a6859555a476542717669626f66494d4d487467",
			"revision": 2,
			"signature": "f3dc30c2255254a7ffd64e767e15f8b9dc908491907c79afb3a1b24ee3b9602f10ff01bce22e1e700f502190fff4ee209f5b32e4c2b9e1ef6b0bed0c2b558406"
		}`)

	opts := skynet.DefaultUploadOptions
	gock.New(skynet.DefaultPortalURL()).
		Post(opts.EndpointPath).
		Reply(200).
		JSON(map[string]string{"skylink": skylink})

	gock.New(skynet.DefaultPortalURL()).
		Post(skynet.RegistryEndpoint).
		Reply(500). // server error during SetEntry
		BodyString("test")

	err := client.SetJSON(
		"191907ba9a408d2851cd0fa68c2f264ec8924d0d6ff64fc858070ee3b77b25524a964fa1cb329d066aedcf7fc03a249eeea3cf2461811090b287daaaec37ab36",
		"TEST_KEY",
		bytes.NewBuffer([]byte("test2")),
		nil,
	)

	// Check that the content type is set.
	if !strings.Contains(interceptedRequest, `"revision":3`) {
		t.Fatal("incorrect revision bump")
	}

	// Verify we don't have pending mocks.
	if !gock.IsDone() {
		t.Fatal("test finished with pending mocks")
	}

	if err == nil {
		t.Fatal(err)
	}
}
