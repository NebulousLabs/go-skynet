package tests

import (
	"reflect"
	"testing"

	skynet "github.com/NebulousLabs/go-skynet"
	"gopkg.in/h2non/gock.v1"
)

// TestAddSkykey tests adding a skykey.
func TestAddSkykey(t *testing.T) {
	defer gock.Off()

	const skykey = "skykey:BAAAAAAAAABrZXkxAAAAAAAAAAQgAAAAAAAAADiObVg49-0juJ8udAx4qMW-TEHgDxfjA0fjJSNBuJ4a"

	opts := skynet.DefaultAddSkykeyOptions
	gock.New(skynet.DefaultPortalURL).
		Post(opts.PortalAddSkykeyPath).
		MatchParam("skykey", skykey).
		Reply(200)

	err := skynet.AddSkykey(skykey, opts)
	if err != nil {
		t.Fatal(err)
	}

	// Verify we don't have pending mocks.
	if !gock.IsDone() {
		t.Fatal("test finished with pending mocks")
	}
}

// TestCreateSkykey tests creating a skykey.
func TestCreateSkykey(t *testing.T) {
	defer gock.Off()

	const skykey = "skykey:BAAAAAAAAABrZXkxAAAAAAAAAAQgAAAAAAAAADiObVg49-0juJ8udAx4qMW-TEHgDxfjA0fjJSNBuJ4a"
	const name = "testcreateskykey"
	const id = "pJAPPfWkWXpss3BvMDCJCw=="
	const skykeyType = "private-id"

	opts := skynet.DefaultCreateSkykeyOptions
	gock.New(skynet.DefaultPortalURL).
		Post(opts.PortalCreateSkykeyPath).
		MatchParam("name", name).
		MatchParam("type", skykeyType).
		Reply(200).
		JSON(skynet.Skykey{Skykey: skykey, Name: name, ID: id, Type: skykeyType})

	skykey2, err := skynet.CreateSkykey(name, skykeyType, opts)
	if err != nil {
		t.Fatal(err)
	}

	expectedSkykey := skynet.Skykey{
		Skykey: skykey,
		Name:   name,
		ID:     id,
		Type:   skykeyType,
	}
	if !reflect.DeepEqual(expectedSkykey, skykey2) {
		t.Fatalf("expected skykey %v, got %v", expectedSkykey, skykey2)
	}

	// Verify we don't have pending mocks.
	if !gock.IsDone() {
		t.Fatal("test finished with pending mocks")
	}
}

// TestGetSkykey tests getting a skykey by name and by ID.
func TestGetSkykey(t *testing.T) {
	defer gock.Off()

	const name = "testcreateskykey"
	const id = "pJAPPfWkWXpss3BvMDCJCw=="

	// Verify we don't have pending mocks.
	if !gock.IsDone() {
		t.Fatal("test finished with pending mocks")
	}
}

// TestListSkykeys tests listing skykeys.
func TestListSkykeys(t *testing.T) {
	defer gock.Off()

	skykey1 := skynet.Skykey{
		Skykey: "skykey:BAAAAAAAAABrZXkxAAAAAAAAAAQgAAAAAAAAADiObVg49-0juJ8udAx4qMW-TEHgDxfjA0fjJSNBuJ4a",
		Name:   "skykey1",
		ID:     "id1",
	}
	skykey2 := skynet.Skykey{
		Skykey: "skykey:BAAAAAAAAABrZXkxAAAAAAAAAAQgAAAAAAAAADiObVg49-0juJ8udAx4qMW-TEHgDxfjA0fjJSNBuJ4a",
		Name:   "skykey2",
		ID:     "id2",
	}
	response := []skynet.Skykey{skykey1, skykey2}

	opts := skynet.DefaultListSkykeysOptions
	gock.New(skynet.DefaultPortalURL).
		Get(opts.PortalListSkykeysPath).
		Reply(200).
		JSON(map[string][]skynet.Skykey{"skykeys": response})

	skykeys, err := skynet.ListSkykeys(opts)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(skykeys, response) {
		t.Fatalf("expected %v, got %v", response, skykeys)
	}

	// Verify we don't have pending mocks.
	if !gock.IsDone() {
		t.Fatal("test finished with pending mocks")
	}
}
