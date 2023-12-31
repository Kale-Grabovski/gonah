FROM golang:1.21.0 as base

WORKDIR /tmp/gonah

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o gonah .

#FROM gcr.io/distroless/static-debian11
FROM debian:bookworm-slim

COPY --from=base /tmp/gonah/gonah .
COPY --from=base /tmp/gonah/migrations/ migrations/
COPY --from=base /tmp/gonah/config-example.yaml .

CMD ["./gonah", "api"]
