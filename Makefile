VERSION ?= 0.1.0
BIN=heravision

build:
	CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(BIN) ./cmd/heravision

test:
	go test ./... -count=1

vet:
	go vet ./...

lint:
	golangci-lint run ./...

bench:
	./$(BIN) bench --n 20

install:
	go install ./cmd/heravision

clean:
	rm -f $(BIN) $(BIN).exe
