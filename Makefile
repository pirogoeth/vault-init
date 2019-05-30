.PHONY: build

build:
	go build -o vault-init ./cmd/vault-init/...

test:
	go test -v ./...

docker:
	docker build --no-cache --pull -t containers.dev.maio.me/pirogoeth/vault-init:latest -f Dockerfile .
	docker build --no-cache --pull -t containers.dev.maio.me/pirogoeth/vault-init:debian-latest -f Dockerfile.debian .
	docker push containers.dev.maio.me/pirogoeth/vault-init:latest
	docker push containers.dev.maio.me/pirogoeth/vault-init:debian-latest
