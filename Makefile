GO ?= go
BUF ?= buf
STATICCHECK ?= $(GO) run honnef.co/go/tools/cmd/staticcheck@2026.2.1
GO_LICENSES_VERSION ?= v2.0.1
ADDLICENSE ?= $(GO) run github.com/google/addlicense@v1.2.0 -c "Joshua Winkler and The yolocam Authors" -l apache
HEADER_FILES = $$(git ls-files '*.go' '*.proto' | grep -v '^internal/genproto/')
LICENSES_DIR ?= third_party_licenses

.PHONY: all build test vet fmt fmt-check staticcheck tidy generate licenses license-headers license-headers-check lint format breaking check

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

# GOOS=windows just because the Windows build has the most dependencies (mousetrap), thus a superset of the others.
licenses:
	$(GO) install github.com/google/go-licenses/v2@$(GO_LICENSES_VERSION)
	GOOS=windows "$$($(GO) env GOPATH)/bin/go-licenses" save ./cmd/yolocam \
		--save_path=$(LICENSES_DIR) --ignore github.com/bemoty/yolocam --force
	mkdir -p $(LICENSES_DIR)/go
	# Distro packages (e.g. Arch) move Go's LICENSE out of GOROOT.
	cp "$$($(GO) env GOROOT)/LICENSE" $(LICENSES_DIR)/go/LICENSE 2>/dev/null || \
		cp /usr/share/licenses/go/LICENSE $(LICENSES_DIR)/go/LICENSE

license-headers:
	$(ADDLICENSE) $(HEADER_FILES)

license-headers-check:
	$(ADDLICENSE) -check $(HEADER_FILES)

lint:
	$(BUF) lint

format:
	$(BUF) format -w

breaking:
	$(BUF) breaking --against '.git#branch=main'

check: fmt-check license-headers-check vet staticcheck lint breaking test
