FROM golang:alpine AS build

WORKDIR /source
ADD . /source/
RUN apk add --no-cache git make && \
    make build

# ---
FROM alpine:latest
LABEL maintainer="Sean Johnson <me@seanj.dev>"

COPY --from=build /source/vault-init /bin/vault-init

CMD ["vault-init"]