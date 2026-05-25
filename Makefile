BINARY   := lumina
CMD      := ./cmd/lumina
VERSION  ?= dev
MODULE   := github.com/kaduvelasco/lumina-tools
LDFLAGS  := -X $(MODULE)/internal/version.Version=$(VERSION)
DIST     := dist

.PHONY: build run test lint clean install release

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(CMD)

run:
	go run $(CMD) $(ARGS)

test:
	go test -race ./...

lint:
	go vet ./...

clean:
	rm -f $(BINARY)
	rm -rf $(DIST)

install: build
	sudo install -m 755 $(BINARY) /usr/local/bin/$(BINARY)

release:
	@echo "Building release $(VERSION)..."
	@mkdir -p $(DIST)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/lumina-linux-amd64 $(CMD)
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/lumina-linux-arm64 $(CMD)
	@echo "Binaries in $(DIST):"
	@ls -lh $(DIST)/
