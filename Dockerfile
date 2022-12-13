FROM golang:1.19-alpine as builder

RUN set -x \
    && apk add --no-cache git make curl
COPY . /build/app/
WORKDIR /build/app

ARG APP_VERSION
ARG APP_BIN="metachart"
ARG CI_USER
ARG CI_TOKEN

ARG UPX_COMPRESSION="yes"
ARG UPX_VERSION="3.96"
RUN set -x \
  && if [ "$(uname -m)" = "aarch64" ]; then \
    export UPX_ARCH="arm64"; \
  else \
    export UPX_ARCH="amd64"; \
  fi \
  && curl -O -L https://github.com/upx/upx/releases/download/v${UPX_VERSION}/upx-${UPX_VERSION}-${UPX_ARCH}_linux.tar.xz \
  && tar xvJf upx-${UPX_VERSION}-${UPX_ARCH}_linux.tar.xz -C /tmp/ \
  && mv /tmp/upx*/upx /bin/ \
  && rm -rf /tmp/upx*

RUN set -x \
    && CGO_ENABLED=0 \
        go build -ldflags "-X main.appVersion=${APP_VERSION} -w -s" \
                 -o ./bin/${APP_BIN} cmd/main.go \
    && if [ "${UPX_COMPRESSION}" = "yes" ]; then \
        upx -o /build/${APP_BIN} ./bin/${APP_BIN} ; \
    else \
        mv ./bin/${APP_BIN} /build/${APP_BIN}; \
    fi \
    && chmod +x /build/${APP_BIN}

FROM alpine:latest
ARG APP_BIN="metachart"
COPY --from=builder /build/${APP_BIN} /bin/${APP_BIN}
