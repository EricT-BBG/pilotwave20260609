# build npm stage
FROM --platform=$BUILDPLATFORM node:24-bookworm-slim AS npm-builder
ARG APP_VERSION=v1.3
ARG BUILD_TIME=unknown
ENV BUILD_SRC=/pilotwave_src/
RUN mkdir -p ${BUILD_SRC}
COPY . ${BUILD_SRC}
WORKDIR ${BUILD_SRC}
RUN apt-get update \
    && apt-get install -y --no-install-recommends git make ca-certificates \
    && rm -rf /var/lib/apt/lists/*
RUN VITE_APP_VERSION="${APP_VERSION}" VITE_BUILD_TIMESTAMP="${BUILD_TIME}" make build-web

#build go stage
FROM --platform=$BUILDPLATFORM golang:1.25-bookworm AS builder
ARG TARGETOS
ARG TARGETARCH
ARG APP_VERSION=v1.3
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
ENV BUILD_SRC=/pilotwave_src/
COPY --from=npm-builder ${BUILD_SRC} ${BUILD_SRC}
#RUN mkdir -p ${BUILD_SRC}
#COPY . ${BUILD_SRC}
WORKDIR ${BUILD_SRC}
RUN apt-get update \
    && apt-get install -y --no-install-recommends git make gcc gcc-aarch64-linux-gnu gcc-x86-64-linux-gnu libc6-dev-amd64-cross libc6-dev-arm64-cross \
    && rm -rf /var/lib/apt/lists/*
RUN go generate ./cmd/pilotwave 
RUN case "${TARGETARCH}" in \
        amd64) export CC=x86_64-linux-gnu-gcc ;; \
        arm64) export CC=aarch64-linux-gnu-gcc ;; \
        *) echo "unsupported target arch: ${TARGETARCH}" >&2; exit 1 ;; \
    esac \
    && CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
        -ldflags "-X git.brobridge.com/pilotwave/pilotwave/pkg/buildinfo.Version=${APP_VERSION} -X git.brobridge.com/pilotwave/pilotwave/pkg/buildinfo.Commit=${COMMIT} -X git.brobridge.com/pilotwave/pilotwave/pkg/buildinfo.BuildTime=${BUILD_TIME}" \
        -o /pilotwave ./cmd/pilotwave
RUN mkdir -p /runtime/pilotwaveDB \
    && touch /runtime/pilotwaveDB/.keep \
    && chmod 0755 /runtime/pilotwaveDB

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /
COPY --from=builder /pilotwave /pilotwave
COPY ./configs/config.toml.dist /config.toml
COPY --from=builder /pilotwave_src/web/dist /web/dist
COPY --from=builder --chown=nonroot:nonroot /runtime/pilotwaveDB /pilotwaveDB

EXPOSE 22112
ENTRYPOINT ["/pilotwave"]
