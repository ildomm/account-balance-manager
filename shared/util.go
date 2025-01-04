package shared

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"flag"
	"fmt"
	"net/url"
)

const (
	DefaultListenAddress = 8080
)

var (
	// Signals that we will handle
	signals = []os.Signal{syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT}
)

func WaitForSignal() os.Signal {
	// Catch signals
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, signals...)

	// Wait for a signal to exit
	signal := <-sigchan
	return signal
}

func SetGlobalTimezoneUTC() error {
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		return err
	}
	time.Local = loc
	return nil
}

func ParseDBConnURL(args []string) (string, error) {
	var dBConnURL string

	fs := flag.FlagSet{}
	fs.StringVar(&dBConnURL, "db", os.Getenv("DATABASE_URL"), "Postgres connection URL, eg: postgres://user:pass@host:5432/dbname. Must be a valid URL.")

	err := fs.Parse(args)
	if err != nil {
		return "", err
	}

	// Postgres URLs follow this form: postgres://user:password@host:port/dbname?args
	// Parse them as a URL to ensure they are valid, otherwise return an error.
	_, err = url.Parse(dBConnURL)
	if err != nil {
		return "", fmt.Errorf("the -db or DATABASE_URL url is not valid")
	}

	if dBConnURL == "" {
		return "", fmt.Errorf("missing -db or DATABASE_URL")
	}

	return dBConnURL, nil
}

func ParseHTTPPort(args []string) (int, error) {
	var httpServerPort int

	fs := flag.FlagSet{}
	fs.IntVar(
		&httpServerPort,
		"http-server-port",
		DefaultListenAddress,
		fmt.Sprintf("The http server port to listen on for incoming API requests, eg: '8080', defaults to %d", DefaultListenAddress),
	)

	err := fs.Parse(args)
	if err != nil {
		return 0, err
	}

	return httpServerPort, nil
}
