VERSION ?= 0.1.1
BIN=heravision
LDFLAGS=-s -w -X heravision/internal/buildinfo.Version=$(VERSION)

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/heravision

test:
	go test ./... -count=1

vet:
	go vet ./...

lint:
	golangci-lint run ./...

bench: build
	./$(BIN) bench --n 20

install:
	go install ./cmd/heravision

clean:
	rm -f $(BIN) $(BIN).exe

.PHONY: build test vet lint bench install clean
