# syntax=docker/dockerfile:1
FROM golang:1.20-alpine as builder
WORKDIR /build
ADD . /build/
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -o hook-hub .

FROM alpine
RUN apk add --no-cache
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/hook-hub /app
EXPOSE 3322
ENTRYPOINT ["/app/hook-hub"]
