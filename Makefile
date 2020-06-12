.PHONY: all
all: build

.PHONY: build
build: build-proj cmd

.PHONY: clean
clean:
	rm -rf _tools
	go clean -cache -testcache -modcache

.PHONY: cmd
cmd:
	go build -o bin/peakbagger ./cmd

.PHONY: build-proj
build-proj:
	go build ./...

.PHONY: fmt
fmt:
	@echo "==> running Go format <=="
	gofmt -s -l -w .

.PHONY: lint
lint:
	@echo "==> linting Go code <=="
	golangci-lint run ./...
	@echo "==> running go vet <=="
	go vet ./...

.PHONY: test
test:
	@echo "==> running Go tests <=="
	CGO_ENABLED=1 go test -p 1 -race ./...
