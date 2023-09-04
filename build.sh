#!/bin/sh

export GOFLAGS="-mod=vendor"

go build -o bin/piska main.go
