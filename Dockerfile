# Build the manager binary
FROM golang:1.13.4 as builder

WORKDIR /workspace
# Copy the Go Modules manifests

COPY go.mod go.mod
COPY go.sum go.sum

ENV CGO_ENABLED=0 \
    GOOS="linux" \
    GO_APP_PKG="github.com/goharbor/harbor-operator" \
    GO111MODULE=on

# Copy the go source
COPY main.go main.go
COPY api api
COPY pkg pkg
COPY controllers controllers
COPY assets assets

COPY hack hack

COPY Makefile Makefile

RUN make generate

COPY vendor vendor
# Build
RUN go build -a \
    -ldflags "-X ${GO_APP_PKG}.OperatorVersion=${RELEASE_VERSION}" \
    -o manager \
    pkged.go \
    main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER nonroot:nonroot

ENTRYPOINT ["/manager"]
