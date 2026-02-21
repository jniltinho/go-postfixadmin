# Stage 1: Build CSS with Tailwind
FROM node:20-alpine AS css-builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npx @tailwindcss/cli -i ./public/css/input.css -o ./public/css/style.css

# Stage 2: Build the Go application
FROM golang:1.25-alpine AS go-builder
RUN apk add --no-cache upx
WORKDIR /app
# Copy dependency manifests
COPY go.mod go.sum ./
RUN go mod download
# Copy the rest of the source code
COPY . .
# Copy the generated CSS from the previous stage
COPY --from=css-builder /app/public/css/style.css ./public/css/style.css
# Build the binary
RUN CGO_ENABLED=0 go build -o postfixadmin -v -ldflags="-s -w"
# Compress the binary
RUN upx postfixadmin

# Stage 3: Final minimal image
FROM alpine:3.21
ENV TZ=America/Sao_Paulo
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
# Copy the binary from the builder stage
COPY --from=go-builder /app/postfixadmin .
COPY config.toml.example config.toml
# Expose the default port
EXPOSE 8080
# Set the entrypoint
ENTRYPOINT ["/app/postfixadmin"]
# Default command starting the server
CMD ["server", "--port", "8080"]
