FROM golang:alpine3.19 AS builder

RUN apk --no-cache add ca-certificates make

WORKDIR /build
ADD . .
RUN go build -ldflags="-X main.serviceVersion=v0.0.0" -o ./bin/test-service .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/bin/* .