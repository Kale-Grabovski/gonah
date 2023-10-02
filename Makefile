all: build

build:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build

test:
	make build && go clean -testcache && go test ./...
