# Stage 1: Build Vue 3 frontend
FROM --platform=$BUILDPLATFORM node:24-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
COPY frontend/vendor ./vendor
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi
COPY frontend/ .
RUN npm run build

# Stage 2: Build the Go application
# Go cross-compiles on the build host (no QEMU for the slow part); TARGETOS/TARGETARCH come from buildx
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS go-builder
# no defaults: a default would override the value buildx sets for each platform
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ENV TZ=America/Sao_Paulo
RUN apk add --no-cache upx tzdata
WORKDIR /app
COPY go.mod go.sum .
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/web/dist ./web/dist
RUN BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) && \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o bin/postfixadmin \
    -ldflags "-s -w \
    -X go-postfixadmin/cmd.Version=${VERSION} \
    -X go-postfixadmin/cmd.BuildDate=${BUILD_DATE} \
    -X go-postfixadmin/cmd.GitCommit=${GIT_COMMIT}" && \
    upx --best --lzma bin/postfixadmin


FROM go-builder AS test-runner
RUN apk add --no-cache gcc musl-dev
ENV CGO_ENABLED=1
CMD ["go", "test", "-race", "-tags", "integration", "./..."]

# Stage 3: Final minimal image
FROM alpine:3.21
LABEL org.opencontainers.image.source="https://github.com/jniltinho/go-postfixadmin" \
      org.opencontainers.image.description="Go-PostfixAdmin: web panel and REST API for Postfix + Dovecot mail servers" \
      org.opencontainers.image.licenses="MIT"
ENV TZ=America/Sao_Paulo
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /app/bin/postfixadmin .
COPY config.toml.example config.toml
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

EXPOSE 8080
ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["./postfixadmin", "server"]
