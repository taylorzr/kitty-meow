install:
  go install .

build:
  go build -ldflags "-X main.version=$(git describe --tags --always) -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" .

test:
  go test ./...

fmt:
  gofmt -w .

lint:
  golangci-lint run

check: lint test

clean:
  rm -f kitty-meow meow
