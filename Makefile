.PHONY: clean build

.PHONY: clean
clean:
	rm -rf build
	mkdir -p build

.PHONY: deps
deps:
	go env -w "GOPRIVATE=github.com/ildomm/*"
	go mod download

.PHONY: unit-test
unit-test: deps
	go test -tags=testing -count=1 ./...