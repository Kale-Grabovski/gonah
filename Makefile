all: build

build:
	GOOS=linux GOARCH=amd64 go build

test:
	make build && go test ./...
