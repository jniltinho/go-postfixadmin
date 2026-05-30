# Stage 1: Build Vue 3 frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
COPY frontend/vendor ./vendor
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Build the Go application
FROM golang:1.26-alpine AS go-builder
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
    CGO_ENABLED=0 go build -o bin/postfixadmin \
    -ldflags "-s -w \
    -X go-postfixadmin/cmd.Version=${VERSION} \
    -X go-postfixadmin/cmd.BuildDate=${BUILD_DATE} \
    -X go-postfixadmin/cmd.GitCommit=${GIT_COMMIT}" && \
    upx --best --lzma bin/postfixadmin


# Stage 3: Final minimal image
FROM alpine:3.21
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
