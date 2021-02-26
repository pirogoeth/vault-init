ARG ALPINE_REPO=containers.dev.maio.me/library/alpine
ARG ALPINE_VERSION=v3.13
ARG GOLANG_REPO=containers.dev.maio.me/library/golang
ARG GOLANG_VERSION=1.15.6

FROM ${GOLANG_REPO}:${GOLANG_VERSION} as build

ADD . /source
WORKDIR /source
RUN apk add --no-cache bash git make && \
    make build out=/source/vault-init

# ---

FROM ${ALPINE_REPO}:${ALPINE_VERSION}

LABEL maintainer="Sean Johnson <me@seanj.dev>"

COPY --from=build /source/vault-init /bin/vault-init

CMD ["vault-init"]