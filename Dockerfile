# syntax=docker/dockerfile:1.7
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/entropy-shear ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S es && adduser -S -G es es && \
    mkdir -p /app/ledger && chown -R es:es /app
WORKDIR /app
COPY --from=build /out/entropy-shear /app/entropy-shear
COPY --from=build /src/examples /app/examples
USER es
ENV ENTROPY_SHEAR_ADDR=":8080" \
    ENTROPY_SHEAR_LEDGER="/app/ledger/shear-chain.jsonl"
EXPOSE 8080
ENTRYPOINT ["/app/entropy-shear"]
