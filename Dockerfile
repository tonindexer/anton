# build emulator libraries
FROM alpine:3 AS emulator-builder

RUN apk add --no-cache git

RUN git clone --recurse-submodules -b emulator_vm_verbosity https://github.com/dungeon-master-666/ton.git

RUN apk add --no-cache make cmake gcc g++ musl-dev zlib-dev openssl-dev linux-headers

RUN mkdir build && (cd build && cmake ../ton -DCMAKE_BUILD_TYPE=Release && cmake --build . --target emulator)
RUN mkdir /output && cp build/emulator/libemulator.so /output


# build
FROM golang:1.19-alpine AS builder

RUN apk add --no-cache build-base

#prepare env
WORKDIR /go/src/github.com/tonindexer/anton

RUN go install github.com/swaggo/swag/cmd/swag@v1.8.10

# download dependencies
COPY go.mod go.sum /go/src/github.com/tonindexer/anton/
RUN go mod download

# copy application code
COPY migrations /go/src/github.com/tonindexer/anton/migrations
COPY cmd /go/src/github.com/tonindexer/anton/cmd
COPY addr /go/src/github.com/tonindexer/anton/addr
COPY abi /go/src/github.com/tonindexer/anton/abi
COPY internal /go/src/github.com/tonindexer/anton/internal
COPY main.go /go/src/github.com/tonindexer/anton

# copy emulator library to the tongo package, which uses it
COPY --from=emulator-builder /output/libemulator.so /go/pkg/mod/github.com/tonkeeper/tongo@v1.0.14/lib/linux/libemulator.so

RUN swag init \
        --output api/http --generalInfo internal/api/http/controller.go \
        --parseDependency --parseInternal

RUN go build -o /anton /go/src/github.com/tonindexer/anton


# application
FROM alpine:3

ENV LISTEN=0.0.0.0:8080

RUN apk add --no-cache libgcc libstdc++

RUN addgroup -S anton && adduser -S anton -G anton
WORKDIR /app
COPY --from=emulator-builder /output/libemulator.so /lib
COPY --from=builder /anton /usr/bin/anton

USER anton:anton
EXPOSE 8080
ENTRYPOINT ["/usr/bin/anton"]
