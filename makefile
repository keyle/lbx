BINARY := lbx
BIN_DIR := bin

.PHONY: all build release test lint fmt vet run clean

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

install: release
	@mkdir -p /usr/local/bin
	install -m755 bin/lbx /usr/local/bin/.

run:
	go run ./cmd/lbx

clean:
	rm -f $(BINARY)
	rm -rf $(BIN_DIR)
