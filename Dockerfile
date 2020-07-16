# Build the manager binary
FROM golang:1.14.5 as builder

WORKDIR /workspace
# Copy the Go Modules manifests

COPY go.mod go.mod
COPY go.sum go.sum

ENV CGO_ENABLED=0 \
    GOOS="linux" \
    GO_APP_PKG="github.com/goharbor/harbor-operator" \
    GO111MODULE=on

# Copy the go source
COPY . .

RUN make manager

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/bin/manager .
USER nonroot:nonroot

ENTRYPOINT ["/manager"]
