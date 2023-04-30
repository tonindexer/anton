# build
FROM --platform=linux/amd64 golang:1.19 AS build

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

RUN swag init \
        --output api/http --generalInfo internal/api/http/controller.go \
        --parseDependency --parseInternal && \
    go build -o /anton /go/src/github.com/tonindexer/anton


# application
FROM --platform=linux/amd64 debian

ENV LISTEN=0.0.0.0:8080

RUN addgroup --system anton && adduser --system anton
WORKDIR /app
COPY --from=build /go/pkg/mod/github.com/tonkeeper/tongo@v1.0.14/lib/linux/libemulator.so /lib
COPY --from=build /anton /usr/bin/anton

USER anton:anton
EXPOSE 8080
ENTRYPOINT ["/usr/bin/anton"]
