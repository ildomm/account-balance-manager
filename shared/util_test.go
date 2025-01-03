package shared

import (
	"github.com/stretchr/testify/require"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestWaitForSignal(t *testing.T) {
	// Use a buffered channel to avoid blocking the sender goroutine
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	go func() {
		// Simulate sending a signal after a delay
		time.Sleep(200 * time.Millisecond)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	// Add a delay to allow the signal handler to execute
	time.Sleep(100 * time.Millisecond)

	signalReceived := WaitForSignal()

	if signalReceived != syscall.SIGINT {
		t.Errorf("Expected signal SIGINT, got %v", signalReceived)
	}
}

func TestMissingDbURL(t *testing.T) {
	_, err := ParseDBConnURL([]string{})
	if err == nil || err.Error() != "missing -db or DATABASE_URL" {
		t.Fatalf("Wrong error, got %v", err)
	}
}

func TestInvalidDbURL(t *testing.T) {
	_, err := ParseDBConnURL([]string{
		"-db",
		"postgres://user:pass@host:port-not-a-number/dbname2"})
	if err == nil || err.Error() != "the -db or DATABASE_URL url is not valid" {
		t.Fatalf("Wrong error, got %v", err)
	}
}

func TestSetGlobalTimezoneUTC(t *testing.T) {
	err := SetGlobalTimezoneUTC()
	require.NoError(t, err)

	// Check if time.Local is set to UTC
	require.Equal(t, time.UTC, time.Local)

	// Optional: Test some time-related functions
	now := time.Now()
	require.Equal(t, "UTC", now.Location().String())
}

func TestValidDbURL(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@host:5432/dbname")
	defer os.Unsetenv("DATABASE_URL")

	url, err := ParseDBConnURL([]string{})
	require.NoError(t, err)
	require.Equal(t, "postgres://user:pass@host:5432/dbname", url)
}

func TestParseHTTPPortDefault(t *testing.T) {
	port, err := ParseHTTPPort([]string{})
	require.NoError(t, err)
	require.Equal(t, DefaultListenAddress, port)
}

func TestParseHTTPPortCustom(t *testing.T) {
	args := []string{"-http-server-port", "8080"}
	port, err := ParseHTTPPort(args)
	require.NoError(t, err)
	require.Equal(t, 8080, port)
}

func TestParseHTTPPortInvalidPort(t *testing.T) {
	args := []string{"-http-server-port", "notaport"}
	_, err := ParseHTTPPort(args)
	require.Error(t, err)
}
