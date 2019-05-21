FROM golang:alpine AS build

RUN apk add --no-cache git && \
    go get -v glow.dev.maio.me/seanj/vault-init/cmd/vault-init

# ---
FROM alpine:latest
LABEL maintainer="Sean Johnson <me@seanj.dev>"

COPY --from=build /go/bin/vault-init /bin/vault-init

CMD ["vault-init"]