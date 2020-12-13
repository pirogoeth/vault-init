SHELL := $(shell which bash)
.PHONY: build cross docker test test/integration test/integration/clean

build:
	go build -o vault-init ./cmd/vault-init/...

cross:
	mkdir -p release
	env GOOS=linux GOARCH=amd64 go build -o release/vault-init_linux_amd64 ./cmd/vault-init/...
	env GOOS=linux GOARCH=arm GOARM=7 go build -o release/vault-init_linux_armv7 ./cmd/vault-init/...
	env GOOS=linux GOARCH=arm64 go build -o release/vault-init_linux_arm64 ./cmd/vault-init/...
	env GOOS=linux GOARCH=386 go build -o release/vault-init_linux_386 ./cmd/vault-init/...
	env GOOS=darwin GOARCH=amd64 go build -o release/vault-init_darwin_amd64 ./cmd/vault-init/...
	ls ./release/* | xargs -I{} tar czvpf {}.tar.gz {}

test:
	go test -v ./...

docker:
	docker build --no-cache --pull -t containers.dev.maio.me/pirogoeth/vault-init:latest -f Dockerfile .
	docker build --no-cache --pull -t containers.dev.maio.me/pirogoeth/vault-init:debian-latest -f Dockerfile.debian .
	docker push containers.dev.maio.me/pirogoeth/vault-init:latest
	docker push containers.dev.maio.me/pirogoeth/vault-init:debian-latest

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
