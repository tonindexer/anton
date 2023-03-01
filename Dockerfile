# build
FROM golang:1.18-alpine AS build

#prepare env
WORKDIR /go/src/github.com/iam047801/tonidx
RUN apk add --no-cache gcc musl-dev linux-headers

RUN go install github.com/swaggo/swag/cmd/swag@v1.8.10

# download dependencies
COPY go.mod go.sum /go/src/github.com/iam047801/tonidx/
RUN go mod download

# copy application code
COPY internal /go/src/github.com/iam047801/tonidx/internal
COPY cmd /go/src/github.com/iam047801/tonidx/cmd
#COPY api /go/src/github.com/iam047801/tonidx/api
COPY abi /go/src/github.com/iam047801/tonidx/abi
COPY main.go /go/src/github.com/iam047801/tonidx

RUN swag init \
    --output api/http --generalInfo internal/api/http/controller.go \
    --parseDependency --parseInternal

# compile application
RUN go build -o /tonidx /go/src/github.com/iam047801/tonidx


# application
FROM alpine:3

ENV LISTEN=0.0.0.0:80

RUN addgroup -S tonidx && adduser -S tonidx -G tonidx
WORKDIR /app
RUN apk add --no-cache tzdata
COPY --from=build /tonidx /usr/bin/tonidx

USER tonidx:tonidx
EXPOSE 80
ENTRYPOINT ["/usr/bin/tonidx"]
CMD ["indexer"]
