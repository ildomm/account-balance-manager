package server

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewServer tests the NewServer factory function.
func TestNewServer(t *testing.T) {
	server := NewServer()

	assert.Equal(t, DefaultListenAddress, server.ListenAddress())
	assert.Equal(t, DefaultReadHeaderTimeout, server.readHeaderTimeout)
}

// TestServerConfigurationSetters tests the configuration setters.
func TestServerConfigurationSetters(t *testing.T) {
	server := NewServer()

	// Test each setter
	server.WithListenAddress(9090)
	assert.Equal(t, 9090, server.ListenAddress())

	server.WithReadHeaderTimeout(time.Second * 20)
	assert.Equal(t, time.Second*20, server.readHeaderTimeout)

	server.WithWriteTimeout(time.Second * 20)
	assert.Equal(t, time.Second*20, server.writeTimeout)

	server.WithReadTimeout(time.Second * 20)
	assert.Equal(t, time.Second*20, server.readTimeout)

	server.WithIdleTimeout(time.Second * 20)
	assert.Equal(t, time.Second*20, server.idleTimeout)
}

// TestServerRun tests the Run method of the server.
func TestServerRun(t *testing.T) {
	server := NewServer()
	port := rand.Intn(1000) + 8000
	server.WithListenAddress(port)

	go func() {
		err := server.Run()
		assert.NoError(t, err, "server failed to run")
	}()

	// Send a request to the server
	time.Sleep(1 * time.Second) // Wait a moment for the server to start

	// Create a new request with the source-type header
	url := fmt.Sprintf("http://localhost:%d/health", port)
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err, "failed to create request")

	// Use http.DefaultClient to send the request
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err, "request to server failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status code from health check")
}
