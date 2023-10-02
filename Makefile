all: build

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build

test:
	make build && go test ./...
