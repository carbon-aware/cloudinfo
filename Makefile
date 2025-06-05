.PHONY: all test lint clean coverage

all: test lint

test:
	go test -v -race ./...

lint:
	revive run

clean:
	go clean
	rm -f coverage.txt

coverage:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt

# Development tools
tools:
	go install github.com/mgechev/revive@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/onsi/ginkgo/v2/ginkgo@latest 