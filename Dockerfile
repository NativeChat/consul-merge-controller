# Build the consul-merge-controller binary
FROM golang:1.17 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download
COPY Makefile Makefile
RUN make go-mod-vendor-hack GO_MOD_DEPS_DIR=/go/pkg/mod CONSUL_K8S_VERSION=@v0.26.0

# Copy the go source
COPY main.go main.go
COPY apis/ apis/
COPY controllers/ controllers/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o consul-merge-controller main.go

# Use distroless as minimal base image to package the consul-merge-controller binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/consul-merge-controller .
USER 65532:65532

ENTRYPOINT ["/consul-merge-controller"]
