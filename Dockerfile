# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app/financebot ./cmd
RUN mkdir -p /out/app/data
RUN cp /etc/ssl/certs/ca-certificates.crt /out/app/ca-certificates.crt

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
WORKDIR /app
COPY --from=build /out/app/ .

ENV SSL_CERT_FILE=/app/ca-certificates.crt
ENV DATABASE_PATH=/app/data/financebot.db
VOLUME ["/app/data"]

ENTRYPOINT ["/app/financebot"]
