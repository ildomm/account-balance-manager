package server

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"net"

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

	port, err := getAvailablePort(8000)
	assert.NoError(t, err)
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

// getAvailablePort finds an available port starting from the given port
func getAvailablePort(startPort int) (int, error) {
	maxAttempts := 100 // Try up to 100 ports
	for port := startPort; port < startPort+maxAttempts; port++ {
		addr := fmt.Sprintf(":%d", port)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			continue
		}
		ln.Close()
		return port, nil
	}
	return 0, fmt.Errorf("no available ports found in range %d-%d", startPort, startPort+maxAttempts-1)
}
