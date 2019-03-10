package rendertron_test

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"k8s.io/test-infra/prow/plugins/testgrid-screenshot/internal/screenshot"
	"k8s.io/test-infra/prow/plugins/testgrid-screenshot/internal/screenshot/rendertron"
)

func TestSomething(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	client, serverCloser := HTTPTestClient(handler)
	defer serverCloser()

	u := "http://www.google.com"
	w := &bytes.Buffer{}
	o := screenshot.Options{
		Client: client,
	}

	err := rendertron.Capture(u, w, o)

	if err != nil {
		t.Errorf("Expected rendertron not return an error, got: %#v", err)
	}
}

func HTTPTestClient(handler http.Handler) (*http.Client, func()) {
	server := httptest.NewServer(handler)

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, server.Listener.Addr().String())
			},
		},
	}

	return client, server.Close
}
