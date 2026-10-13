BINARY := lbx
BIN_DIR := bin

.PHONY: all build release test lint fmt vet run clean tag install

default: run

build:
	go build -gcflags="all=-N -l" -o $(BINARY) ./cmd/lbx

release: clean
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o $(BIN_DIR)/$(BINARY) ./cmd/lbx

test:
	go test ./...

lint: vet
	@fmt_out="$$(gofmt -l .)"; if [ -n "$$fmt_out" ]; then echo "gofmt needed:"; echo "$$fmt_out"; exit 1; fi

fmt:
	gofmt -w .

vet:
	go vet ./...

tag:
	@current=$$(git describe --tags --abbrev=0 2>/dev/null || echo 0.0.0); \
	next=$$(echo $$current | awk -F. 'BEGIN { OFS = "." } { $$NF = $$NF + 1; print }'); \
	echo "current tag: $$current"; \
	echo "next tag:    $$next"; \
	echo "building..."; \
	go build ./... || { echo "build failed"; exit 1; }; \
	echo "testing..."; \
	go test ./... || { echo "tests failed"; exit 1; }; \
	git tag $$next || exit 1; \
	git push origin $$next || exit 1; \
	echo "tagged and pushed $$next"

install: release
	@mkdir -p /usr/local/bin
	install -m755 bin/lbx /usr/local/bin/.

run:
	go run ./cmd/lbx

clean:
	rm -f $(BINARY)
	rm -rf $(BIN_DIR)
