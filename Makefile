.PHONY: clean build deps unit-test coverage-report coverage-total

# Variables needed when building binaries
VERSION := $(shell grep -oE -m 1 '([0-9]+)\.([0-9]+)\.([0-9]+)' CHANGELOG.md )
GIT_SHA := $(shell git rev-parse HEAD )

.PHONY: clean
clean:
	rm -rf build
	mkdir -p build

.PHONY: deps
deps:
	go env -w "GOPRIVATE=github.com/ildomm/*"
	go mod download

.PHONY: build
build: deps
	# Build the http server  binary
	cd cmd/api && \
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-X main.semVer=${VERSION} -X main.gitSha=${GIT_SHA}" \
        -o ../../build/api

.PHONY: unit-test
unit-test: deps
	go test -tags=testing -count=1 ./...

.PHONY: coverage-report
coverage-report: clean deps
	go test -tags=testing ./... \
		-coverprofile=build/cover.out github.com/ildomm/account-balance-manager/...
	grep -vE 'main\.go|test_helpers' build/cover.out > build/cover.temp && mv build/cover.temp build/cover.out
	go tool cover -html=build/cover.out -o build/coverage.html
	echo "** Coverage is available in build/coverage.html **"

.PHONY: coverage-total
coverage-total: clean deps
	go test -tags=testing ./... \
		-coverprofile=build/cover.out github.com/ildomm/account-balance-manager/...
	grep -vE 'main\.go|test_helpers' build/cover.out > build/cover.temp && mv build/cover.temp build/cover.out
	go tool cover -func=build/cover.out | grep total
