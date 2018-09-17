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

ENTRYPOINT ["/main"]
