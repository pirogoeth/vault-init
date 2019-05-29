.PHONY: build

build:
	go build -o vault-init ./cmd/vault-init/...

test:
	go test -v ./...