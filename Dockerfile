# Dockerfile.distroless
FROM golang:1.21.0-alpine3.18
WORKDIR /
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gonah .
COPY --from=ghcr.io/ufoscout/docker-compose-wait:latest /wait /wait
CMD /wait && /gonah api
