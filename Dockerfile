# build
FROM golang:1.18-alpine AS build

#prepare env
WORKDIR /go/src/github.com/iam047801/tonidx
RUN apk add --no-cache gcc musl-dev linux-headers

# download dependencies
COPY go.mod go.sum /go/src/github.com/iam047801/tonidx/
RUN go mod download

# copy application code
COPY internal /go/src/github.com/iam047801/tonidx/internal
COPY cmd /go/src/github.com/iam047801/tonidx/cmd
COPY main.go /go/src/github.com/iam047801/tonidx

# compile application
RUN go build -o /tonidx /go/src/github.com/iam047801/tonidx


# application
FROM alpine:3

RUN addgroup -S tonidx && adduser -S tonidx -G tonidx
WORKDIR /app
RUN apk add --no-cache tzdata
COPY --from=build /tonidx /usr/bin/tonidx

USER tonidx:tonidx
ENTRYPOINT ["/usr/bin/tonidx"]
CMD ["indexer"]
