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

	const skykey = "testskykey"

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

// TestListSkykeys tests listing skykeys.
func TestListSkykeys(t *testing.T) {
	defer gock.Off()

	skykey1 := skynet.Skykey{
		Skykey: "foo123",
		Name:   "skykey1",
		ID:     "id1",
	}
	skykey2 := skynet.Skykey{
		Skykey: "bar456",
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
