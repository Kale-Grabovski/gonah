# Dockerfile.distroless
FROM golang:1.21.0-alpine3.18 as base

WORKDIR /tmp/gonah

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gonah .

FROM gcr.io/distroless/static-debian11

COPY --from=base /tmp/gonah/gonah .
COPY --from=base /tmp/gonah/config-example.yaml .

CMD ["/gonah", "api"]
