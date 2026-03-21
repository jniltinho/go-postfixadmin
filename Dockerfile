# Stage 1: Build CSS with Tailwind
FROM debian:bookworm AS css-builder
RUN apt-get update && apt install -y curl

WORKDIR /app
RUN curl -ksLO https://github.com/tailwindlabs/tailwindcss/releases/download/v4.2.0/tailwindcss-linux-x64 && \
    chmod +x tailwindcss-linux-x64 && mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss
COPY . .
RUN tailwindcss -i ./web/static/css/input.css -o ./web/static/css/style.css --minify

# Stage 2: Build the Go application
FROM golang:1.26-alpine AS go-builder
ENV TZ=America/Sao_Paulo
RUN apk add --no-cache upx make tzdata
WORKDIR /app
# Copy dependency manifests
COPY go.mod go.sum .
RUN go mod download
# Copy the rest of the source code
COPY . .
# Copy the generated CSS from the previous stage
COPY --from=css-builder /app/public/css/style.css ./public/css/style.css
# Build the binary
RUN make build-docker-prod


# Stage 3: Final minimal image
FROM alpine:3.21
ENV TZ=America/Sao_Paulo
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
# Copy the binary from the builder stage
COPY --from=go-builder /app/bin/postfixadmin .
COPY config.toml.example config.toml
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

# Expose the default port
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]

# Default command starting the server
CMD ["./postfixadmin", "server"]
