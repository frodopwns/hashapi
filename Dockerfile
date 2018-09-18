# STEP 1 build executable binary
FROM golang:alpine as builder
# Install SSL ca certificates
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY . $GOPATH/src/github.com/frodopwns/hashapi/
WORKDIR $GOPATH/src/github.com/frodopwns/hashapi/
#build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/hashapi
# STEP 2 build image
# start from scratch
FROM scratch
COPY --from=builder /go/bin/hashapi /go/bin/hashapi
ENTRYPOINT ["/go/bin/hashapi"]