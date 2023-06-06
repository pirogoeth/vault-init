SHELL := $(shell which bash)

goarch := $(shell go env GOARCH)
goos := $(shell go env GOOS)
out := vault-init

.PHONY: build
build:
	GOOS=$(goos) GOARCH=$(goarch) go build \
		-ldflags "-X github.com/pirogoeth/vault-init/internal/version.Version=$(shell git describe --tags)" \
		-o "$(out)_$(goos)_$(goarch)" \
		./cmd/vault-init/...
	GOOS=$(goos) GOARCH=$(goarch) go build \
		-ldflags "-X github.com/pirogoeth/vault-init/internal/version.Version=$(shell git describe --tags)" \
		-o "$(out)-test-harness_$(goos)_$(goarch)" \
		./cmd/vault-init-test-harness/...


.PHONY: cross
cross:
	mkdir -p release
	make build goos=linux goarch=amd64 out=release/vault-init
	make build goos=linux goarch=arm out=release/vault-init
	make build goos=linux goarch=arm64 out=release/vault-init
	make build goos=linux goarch=386 out=release/vault-init
	make build goos=darwin goarch=amd64 out=release/vault-init
	ls ./release/* | xargs -I{} tar czvpf {}.tar.gz {}

.PHONY: test
test:
	go test -v ./...

.PHONY: docker
docker:
	docker build --no-cache --pull -t ghcr.io/pirogoeth/vault-init:latest -f Dockerfile .
	docker build --no-cache --pull -t ghcr.io/pirogoeth/vault-init:debian-latest -f Dockerfile.debian .
	docker push ghcr.io/pirogoeth/vault-init:latest
	docker push ghcr.io/pirogoeth/vault-init:debian-latest