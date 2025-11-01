# Build the manager binary
FROM golang:1.21.3 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# Cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager cmd/manager/main.go

# Use Red Hat's Universal Base Image (UBI) for runtime
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

# Install required packages
RUN microdnf install -y buildah podman fuse-overlayfs shadow-utils \
    && microdnf clean all \
    && rm -rf /var/cache/yum

# Setup buildah storage configuration
RUN printf '[storage]\ndriver = "overlay"\nrunroot = "/run/containers/storage"\ngraphroot = "/var/lib/containers/storage"\n[storage.options]\nmount_program = "/usr/bin/fuse-overlayfs"\n' > /etc/containers/storage.conf

# Create necessary directories with appropriate permissions
RUN mkdir -p /run/containers/storage /var/lib/containers/storage \
    && chown -R 65532:65532 /run/containers/storage /var/lib/containers/storage

WORKDIR /
COPY --from=builder /workspace/manager .

# Install Trivy
COPY --from=aquasec/trivy:latest /usr/local/bin/trivy /usr/local/bin/trivy

# Set up non-root user
RUN useradd -u 65532 -r -g 0 nonroot \
    && chmod g+w /etc/containers/storage.conf

USER 65532:0

# Set necessary environment variables for rootless buildah
ENV _BUILDAH_STARTED_IN_USERNS="" \
    BUILDAH_ISOLATION=chroot \
    STORAGE_DRIVER=overlay \
    HOME=/home/nonroot

ENTRYPOINT ["/manager"]