BINARY_NAME=http-go-engine
MAIN_PACKAGE_PATH=./main.go

.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: tidy
tidy:
	go fmt ./...
	go vmt ./...
	go mod tidy

.PHONY: audit
audit:
	go mod verify
	go vet ./...
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Skipping..."; \
	fi

.PHONY: run
run:
	go run $(MAIN_PACKAGE_PATH)

.PHONY: test
test:
	go test -v -race -buildvcs ./...

.PHONY: test/cover
test/cover:
	go test -v -race -buildvcs -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

.PHONY: build
build:
	go build -ldflags="-s -w" -o bin/$(BINARY_NAME) $(MAIN_PACKAGE_PATH)

.PHONY: clean
clean:
	go clean
	rm -rf bin/ coverage.out

