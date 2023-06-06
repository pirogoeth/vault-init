goarch := $(shell go env GOARCH)
goos := $(shell go env GOOS)
SHELL := $(shell which bash)
.PHONY: build cross docker test test/integration test/integration/clean

build:
ifdef out
	GOOS=$(goos) GOARCH=$(goarch) go build \
		-ldflags "-X github.com/pirogoeth/vault-init/internal/version.Version=$(shell git describe --tags)" \
		-o "$(out)_$(goos)_$(goarch)" \
		./cmd/vault-init/...
	GOOS=$(goos) GOARCH=$(goarch) go build \
		-ldflags "-X github.com/pirogoeth/vault-init/internal/version.Version=$(shell git describe --tags)" \
		-o "$(out)-test-harness_$(goos)_$(goarch)" \
		./cmd/vault-init-test-harness/...
else
	GOOS=$(goos) GOARCH=$(goarch) go build \
		-ldflags "-X github.com/pirogoeth/vault-init/internal/version.Version=$(shell git describe --tags)" \
		-o "vault-init_$(goos)_$(goarch)" \
		./cmd/vault-init/...
	GOOS=$(goos) GOARCH=$(goarch) go build \
		-ldflags "-X github.com/pirogoeth/vault-init/internal/version.Version=$(shell git describe --tags)" \
		-o "vault-init-test-harness_$(goos)_$(goarch)" \
		./cmd/vault-init-test-harness/...
endif

cross:
	mkdir -p release
	make build goos=linux goarch=amd64 out=release/vault-init_linux_amd64
	make build goos=linux goarch=arm out=release/vault-init_linux_arm
	make build goos=linux goarch=arm64 out=release/vault-init_linux_arm64
	make build goos=linux goarch=386 out=release/vault-init_linux_386
	make build goos=darwin goarch=amd64 out=release/vault-init_darwin_amd64
	ls ./release/* | xargs -I{} tar czvpf {}.tar.gz {}

test:
	go test -v ./...

docker:
	docker build --no-cache --pull -t containers.dev.maio.me/pirogoeth/vault-init:latest -f Dockerfile .
	docker build --no-cache --pull -t containers.dev.maio.me/pirogoeth/vault-init:debian-latest -f Dockerfile.debian .
	docker push containers.dev.maio.me/pirogoeth/vault-init:latest
	docker push containers.dev.maio.me/pirogoeth/vault-init:debian-latest