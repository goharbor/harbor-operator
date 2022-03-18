# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM oamdev/gcr.io-distroless-static:nonroot 
WORKDIR /
COPY manager .
USER nonroot:nonroot

ENTRYPOINT ["/manager"]
