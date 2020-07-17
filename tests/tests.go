package tests

import (
	"net/http"
	"net/http/httputil"

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
