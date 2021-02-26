goarch := $(shell go env GOARCH)
goos := $(shell go env GOOS)
out := vault-init_$(goos)_$(goarch)
.PHONY: build cross docker test test/integration test/integration/clean

build:
	GOOS=$(goos) GOARCH=$(goarch) go build \
		-ldflags "-X glow.dev.maio.me/seanj/vault-init/internal/version.Version=$(shell git describe --tags)" \
		-o "$(out)" \
		./cmd/vault-init/...

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
	docker build --no-cache --pull -t containers.dev.maio.me/seanj/vault-init:latest -f Dockerfile .
	docker build --no-cache --pull -t containers.dev.maio.me/seanj/vault-init:debian-latest -f Dockerfile.debian .
	docker push containers.dev.maio.me/seanj/vault-init:latest
	docker push containers.dev.maio.me/seanj/vault-init:debian-latest

test/integration/clean:
	cd contrib && docker-compose down

test/integration: export VAULT_ADDR = http://localhost:8200
test/integration: test/integration/clean build
	cd contrib && docker-compose up -d
	@echo "Waiting a second, Vault is coming up.."
	@sleep 2
	vault login -method token - <<<"secret"
	vault secrets enable -path /totp totp
	vault write totp/keys/Service generate=true issuer=Vault account_name=vault-init-test
	vault kv put secret/shared session_key=pb5fgEOZwKHf09Zz373a835DteugBmte
	env -i PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin \
		SERVICE_VAULT_API_TOKEN="{{.Vault.token}}" \
		KEY="{{.secret.data.shared.data.session_key}}" \
		OTP="{{.totp.code.Service.code}}" \
		./vault-init \
			--debug \
			--verbose \
			--log-format json \
			--vault-address "http://localhost:8200" \
			--vault-token "secret" \
			--without-reaper \
			--orphan-token \
			--one-shot \
			--path /secret/data/shared \
			--path /totp/code/Service \
			--token-ttl 30s \
			./test.sh
	env -i PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin \
		SERVICE_VAULT_API_TOKEN="{{.Vault.token}}" \
		KEY="{{.secret.data.shared.data.session_key}}" \
		OTP="{{.totp.code.Service.code}}" \
		./vault-init \
			--debug \
			--verbose \
			--log-format json \
			--vault-address "http://localhost:8200" \
			--vault-token "secret" \
			--without-reaper \
			--orphan-token \
			--path /secret/data/shared \
			--path /totp/code/Service \
			--token-ttl 30s \
			./test.sh
