.PHONY: build run clean install

# Nomes dos binários
BINARY_NAME=quota-report

build:
	@echo "Construindo o binário..."
	rm -f $(BINARY_NAME)
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY_NAME) quota-report.go
	@echo "Comprimindo com upx (opcional)..."
	-upx --best --lzma $(BINARY_NAME)

run: build
	@echo "Executando..."
	./$(BINARY_NAME)

install: build
	@echo "Instalando em /usr/local/bin..."
	sudo cp $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	# sudo chown vmail:vmail /usr/local/bin/$(BINARY_NAME)

clean:
	@echo "Limpando..."
	go clean
	rm -f $(BINARY_NAME)
