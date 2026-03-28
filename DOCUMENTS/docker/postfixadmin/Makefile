.PHONY: all build-docker

all: build-docker

build-docker: 
	@echo "Building PostfixAdmin Project..."
	docker build --pull --rm -t postfixadmin-image apache

run: 
	@echo "Starting application..."
	#docker compose up -d
	docker-compose up -d
