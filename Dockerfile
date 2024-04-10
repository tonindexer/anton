# syntax=docker/dockerfile:1.5-labs
FROM debian:12.2-slim AS emulator-builder

ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=Etc/UTC

# build emulator libraries

RUN apt-get update && \
    apt-get install -yqq \
      tzdata build-essential cmake clang openssl \
      libssl-dev zlib1g-dev gperf wget git curl \
      libreadline-dev ccache libmicrohttpd-dev ninja-build pkg-config \
      libsecp256k1-dev libsodium-dev liblz4-dev

ADD --keep-git-dir=true https://github.com/ton-blockchain/ton.git /ton
RUN cd /ton && git submodule update --init --recursive

RUN mkdir build && (cd build && cmake ../ton -DCMAKE_BUILD_TYPE=Release && cmake --build . --target emulator -- -j 8)
RUN mkdir /output && cp build/emulator/libemulator.so /output


# build
FROM golang:1.21.4-bookworm AS builder

RUN apt-get update && \
    apt-get install -y libsecp256k1-1 libsodium23

#prepare env
WORKDIR /go/src/github.com/tonindexer/anton

RUN go install github.com/swaggo/swag/cmd/swag@v1.8.10

# download dependencies
COPY go.mod go.sum /go/src/github.com/tonindexer/anton/
RUN go mod download

# copy application code
COPY migrations /go/src/github.com/tonindexer/anton/migrations
COPY lru /go/src/github.com/tonindexer/anton/lru
COPY cmd /go/src/github.com/tonindexer/anton/cmd
COPY addr /go/src/github.com/tonindexer/anton/addr
COPY abi /go/src/github.com/tonindexer/anton/abi
COPY internal /go/src/github.com/tonindexer/anton/internal
COPY main.go /go/src/github.com/tonindexer/anton

RUN rm /go/pkg/mod/github.com/tonkeeper/tongo@v1.3.0/lib/linux/libemulator.so
COPY --from=emulator-builder /output/libemulator.so /lib/libemulator.so

RUN swag init \
        --output api/http --generalInfo internal/api/http/controller.go \
        --parseDependency --parseInternal

RUN go build -o /anton /go/src/github.com/tonindexer/anton


# application
FROM debian:12.2-slim

ENV LISTEN=0.0.0.0:8080

RUN apt-get update && \
    apt-get install -y libsecp256k1-1 libsodium23 libssl3

RUN groupadd anton && useradd -g anton anton

WORKDIR /app
COPY --from=builder /lib/libemulator.so /lib
COPY --from=builder /go/src/github.com/tonindexer/anton/abi/known /var/anton/known
COPY --from=builder /anton /usr/bin/anton

USER anton:anton
EXPOSE 8080
ENTRYPOINT ["/usr/bin/anton"]
