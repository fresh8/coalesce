FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM fresh8/golang-base:latest as builder
# copy and setup work dir
COPY . /go/src/github.com/fresh8/dredd/circleci-service
WORKDIR /go/src/github.com/fresh8/dredd/circleci-service

# Build executable
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /main

# Minimise binary
RUN strip --strip-unneeded /main
RUN upx /main

FROM scratch
COPY --from=builder /main /main
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["/main"]
