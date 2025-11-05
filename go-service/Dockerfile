# syntax=docker/dockerfile:1

FROM golang:1.25 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=builder /app/server /app/server
# Copy OpenAPI spec to serve via /openapi.yaml
COPY openapi.yaml /app/openapi.yaml
# Expose app port (Fiber default configured via env PORT)
EXPOSE 8080
ENV PORT=8080
USER nonroot
ENTRYPOINT ["/app/server"]