FROM golang:1.23 as build
WORKDIR /workspace
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$(go env GOARCH) go build -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=build /workspace/manager .
USER nonroot:nonroot

ENTRYPOINT ["/manager"]
