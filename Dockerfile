# Stage 1: Build CSS with Tailwind
FROM debian:bookworm AS css-builder
RUN apt-get update && apt install -y curl make

WORKDIR /app
COPY . .
RUN make install-tailwind && make css

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
RUN rm -f ./web/static/css/input.css
# Copy the generated CSS from the previous stage
COPY --from=css-builder /app/web/static/css/style.css ./web/static/css/style.css
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
