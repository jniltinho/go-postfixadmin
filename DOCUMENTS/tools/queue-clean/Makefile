GOCMD ?= go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean -cache
GOGET = $(GOCMD) get
BINARY_NAME = queue-clean


default: get build


get:
	@go mod tidy


build:
	rm -f $(BINARY_NAME)
	CGO_ENABLED=0 $(GOBUILD) -o $(BINARY_NAME) -v -ldflags="-s -w"
	upx $(BINARY_NAME)


build-docker:
	rm -f $(BINARY_NAME)
	docker run --rm -v $(PWD):/app -w /app golang:1.26 $(GOBUILD) -buildvcs=false -o $(BINARY_NAME) -ldflags="-s -w"
	upx $(BINARY_NAME)



clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)


run:
	./$(BINARY_NAME)
