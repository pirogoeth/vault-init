FROM golang:alpine AS build

WORKDIR /source
ADD . /source/
RUN apk add --no-cache git && \
    go install ./cmd/vault-init/...

# ---
FROM alpine:latest
LABEL maintainer="Sean Johnson <me@seanj.dev>"

COPY --from=build /go/bin/vault-init /bin/vault-init

CMD ["vault-init"]