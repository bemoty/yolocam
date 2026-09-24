GO ?= go
BUF ?= buf
STATICCHECK ?= $(GO) run honnef.co/go/tools/cmd/staticcheck@2026.2.1

.PHONY: all build test vet fmt fmt-check staticcheck tidy generate lint format breaking check

all: build test

build:
	$(GO) build ./...

test:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -l -w .

fmt-check:
	@unformatted="$$(gofmt -l .)"; if [ -n "$$unformatted" ]; then echo "$$unformatted"; exit 1; fi

staticcheck:
	$(STATICCHECK) ./...

tidy:
	$(GO) mod tidy

generate:
	$(GO) install tool
	PATH="$$($(GO) env GOPATH)/bin:$$PATH" $(BUF) generate

lint:
	$(BUF) lint

format:
	$(BUF) format -w

breaking:
	$(BUF) breaking --against '.git#branch=main'

check: fmt-check vet staticcheck lint breaking test
