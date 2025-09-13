# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app

# Enable Go modules and download deps
COPY go.mod go.sum ./
RUN go mod download

# Copy the source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /server ./cmd/server

# Runtime stage (distroless with CA certs)
FROM gcr.io/distroless/static-debian12:nonroot
ENV ADDR=:8080
EXPOSE 8080
COPY --from=builder /server /server
USER nonroot:nonroot
ENTRYPOINT ["/server"]

