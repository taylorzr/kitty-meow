alias i := install
alias t := test

install:
  go install .

build:
  go build .

test:
  go test ./...

fmt:
  gofmt -w .

lint:
  golangci-lint run

check: lint test

clean:
  rm -f kitty-meow meow
