# build
FROM golang:1.18-alpine AS build

#prepare env
WORKDIR /go/src/github.com/tonindexer/anton
RUN apk add --no-cache gcc musl-dev linux-headers

RUN go install github.com/swaggo/swag/cmd/swag@v1.8.10

# download dependencies
COPY go.mod go.sum /go/src/github.com/tonindexer/anton/
RUN go mod download

# copy application code
COPY internal /go/src/github.com/tonindexer/anton/internal
COPY cmd /go/src/github.com/tonindexer/anton/cmd
#COPY api /go/src/github.com/tonindexer/anton/api
COPY abi /go/src/github.com/tonindexer/anton/abi
COPY main.go /go/src/github.com/tonindexer/anton

RUN swag init \
    --output api/http --generalInfo internal/api/http/controller.go \
    --parseDependency --parseInternal

# compile application
RUN go build -o /tonidx /go/src/github.com/tonindexer/anton


# application
FROM alpine:3

ENV LISTEN=0.0.0.0:8080

RUN addgroup -S tonidx && adduser -S tonidx -G tonidx
WORKDIR /app
RUN apk add --no-cache tzdata
COPY --from=build /tonidx /usr/bin/tonidx

USER tonidx:tonidx
EXPOSE 8080
ENTRYPOINT ["/usr/bin/tonidx"]
CMD ["indexer"]
