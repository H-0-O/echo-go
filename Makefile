.PHONY: test unit integration build lint

unit:
	go test -count=1 ./...

integration:
	go test -tags=integration -count=1 -timeout=5m ./internal/integration/...

test: unit integration

build:
	go build ./examples/...

lint:
	go vet ./...
